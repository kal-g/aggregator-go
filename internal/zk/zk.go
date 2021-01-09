package zk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-zookeeper/zk"
	"github.com/rs/zerolog"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

// ZkManager handles all zookeeper interaction
type ZkManager struct {
	c                  *zk.Conn
	localOnlyMode      bool
	nodeName           string
	currVote           string
	isLeader           bool
	watchNodesChan     <-chan zk.Event
	nsm                *agg.NamespaceManager
	nodeMap            map[string]bool
	updateMetadataChan <-chan string
}

type AggNodeStatus struct {
	IsReady bool
}

type NodeToNamespaceMapData struct {
	Map map[string]map[string]bool
}

type NamespaceToNodeData struct {
	Node string
}

var logger zerolog.Logger = zerolog.New(os.Stderr).With().
	Str("source", "ZK").
	Timestamp().
	Logger()

// ZkLogger implements logging interface for go-zookeeper package
type ZkLogger struct{}

// Printf implement logger interface
func (l ZkLogger) Printf(fmt string, args ...interface{}) {
	logger.Info().Msgf(fmt, args...)
}

// MakeNewZkManager inits and connects to zk
// If no url given, sets local only mode
func MakeNewZkManager(zkURL string, nodeName string, nsm *agg.NamespaceManager, configFiles []string, updateMetadataChan <-chan string) *ZkManager {
	l := ZkLogger{}
	opt := zk.WithLogger(l)
	c, _, err := zk.Connect([]string{zkURL}, time.Second, opt)

	if err != nil {
		panic(err)
	}

	zkm := &ZkManager{
		c:                  c,
		localOnlyMode:      false,
		nodeName:           nodeName,
		isLeader:           false,
		watchNodesChan:     nil,
		nsm:                nsm,
		nodeMap:            map[string]bool{},
		updateMetadataChan: updateMetadataChan,
	}
	zkm.Setup()
	zkm.LeaderElection()
	go zkm.watchConfigs()
	go zkm.metadataUpdater()
	go zkm.watchMetadata()
	zkm.ingestConfigsToZK(configFiles)
	// TODO Wait for watch nodes init
	time.Sleep(2 * time.Second)
	zkm.DistributeNamespaces()
	return zkm
}

// Setup sets up some directories, making sure they exist
// Also registers the node with the cluster
func (zkm ZkManager) Setup() {
	// Election dir
	_, err := zkm.c.Create("/election", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

	// Nodes directory
	_, err = zkm.c.Create("/nodes", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

	_, err = zkm.c.Create("/namespaceMetadata", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

	// Register Node
	status := AggNodeStatus{
		IsReady: true,
	}
	statusData, err := json.Marshal(status)
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("/nodes/%s", zkm.nodeName)
	_, err = zkm.c.Create(path, statusData, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		if errors.Is(err, zk.ErrNodeExists) {
			logger.Error().Msgf("Node %s already exists in cluster", zkm.nodeName)
			os.Exit(1)
		} else {
			panic(err)
		}
	}

	// Node to namespace map
	nodeMap := NodeToNamespaceMapData{
		Map: map[string]map[string]bool{},
	}
	nodeMapData, _ := json.Marshal(nodeMap)
	_, err = zkm.c.Create("/nodeToNamespaceMap", nodeMapData, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

	_, err = zkm.c.Create("/namespaceToNode", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

	_, err = zkm.c.Create("/configs", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		if !errors.Is(err, zk.ErrNodeExists) {
			panic(err)
		}
	}

}

func (zkm *ZkManager) DistributeNamespaces() {
	masterOnly := false
	if len(zkm.nodeMap) == 1 {
		logger.Info().Msgf("Distributing namespaces to master")
		masterOnly = true
	}
	// Get distributed namespaces
	distributedNs := map[string]string{}
	nss, nstnStat, err := zkm.c.Children("/namespaceToNode")
	if err != nil {
		panic(err)
	}

	// Read each child
	for _, ns := range nss {
		data, _, _ := zkm.c.Get("/namespaceToNode/" + ns)
		nsToNode := NamespaceToNodeData{}
		err = json.Unmarshal(data, &nsToNode)
		if err != nil {
			panic(err)
		}
		distributedNs[ns] = nsToNode.Node
	}

	// Check for nodes that were removed
	data, stat, err := zkm.c.Get("/nodeToNamespaceMap")
	if err != nil {
		panic(err)
	}
	nsmap := NodeToNamespaceMapData{}
	err = json.Unmarshal(data, &nsmap)
	if err != nil {
		logger.Error().Msgf("Error getting nodeToNamespaceMap")
		panic(err)
	}

	// For each node in map that no longer exists, redistribute all namespaces, and delete entry
	for node, namespaces := range nsmap.Map {
		if _, exists := zkm.nodeMap[node]; !exists {
			logger.Info().Msgf("Node %s was removed, redistributing namespaces: %v", node, namespaces)

			for ns := range namespaces {
				delete(distributedNs, ns)
				err = zkm.c.Delete("/namespaceToNode/"+ns, nstnStat.Version)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	// check against metric map to find non distributed namespaces
	nonDistributedNs := []string{}
	for ns := range zkm.nsm.MetricConfigsByNamespace {
		if _, exists := distributedNs[ns]; !exists {
			nonDistributedNs = append(nonDistributedNs, ns)
		}
	}
	logger.Info().Msgf("Namespaces about to be distributed: %+v", nonDistributedNs)

	// Rebuild map, starting with already distributed namespaces
	nsmap.Map = map[string]map[string]bool{}
	logger.Info().Msgf("Distributed NS: %s", distributedNs)
	for ns, node := range distributedNs {
		if _, exists := nsmap.Map[node]; !exists {
			nsmap.Map[node] = map[string]bool{}
		}
		nsmap.Map[node][ns] = true
	}
	// Keep track of changes
	newlyDistributedNamespaces := map[string]string{}
	// Find first non master node
	for c := range zkm.nodeMap {
		if masterOnly || zkm.nodeName != c {
			// Put all non distributed namespaces into map for first non master node
			// TODO Distribute the namespaces evenly
			logger.Info().Msgf("Non distributed NS: %s", nonDistributedNs)
			if _, exists := nsmap.Map[c]; !exists {
				nsmap.Map[c] = map[string]bool{}
			}
			for _, ns := range nonDistributedNs {
				nsmap.Map[c][ns] = true
				newlyDistributedNamespaces[ns] = c
			}
			break
		}
	}
	// Write back to zk
	// Create entries for all new namespaceToNode
	for ns, c := range newlyDistributedNamespaces {
		nsToNode := NamespaceToNodeData{Node: c}
		nsToNodeData, err := json.Marshal(nsToNode)
		if err != nil {
			panic(err)
		}
		_, err = zkm.c.Create("/namespaceToNode/"+ns, nsToNodeData, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			logger.Error().Msgf("Error creating namespaceToNode entry %+v", "/namespaceToNode/"+ns)
			panic(err)
		}
	}
	// Write back map
	logger.Info().Msgf("New NS map: %+v", nsmap.Map)
	data, err = json.Marshal(nsmap)
	if err != nil {
		panic(err)
	}
	_, err = zkm.c.Set("/nodeToNamespaceMap", data, stat.Version)
	if err != nil {
		panic(err)
	}

}

// LeaderElection performs a zk leader election protocol
func (zkm *ZkManager) LeaderElection() {
	vote, err := zkm.c.Create("/election/vote_", []byte(zkm.nodeName), zk.FlagSequence|zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		panic(err)
	}
	tokens := strings.Split(vote, "/")
	zkm.currVote = tokens[len(tokens)-1]
	logger.Info().Msgf("Submitted vote as %s", zkm.currVote)

	// Check election results
	votes, _, err := zkm.c.Children("/election")
	if err != nil {
		logger.Error().Msgf("Error getting votes")
		panic(err)
	}
	sort.Strings(votes)
	if votes[0] == zkm.currVote {
		// If smallest vote, become leader
		logger.Info().Msgf("Became leader")
		zkm.isLeader = true
		// Unassign the nodes owned by this node, since it became leader
		// They will be redistributed in watchNodes
		if len(zkm.nodeMap) > 1 {
			logger.Info().Msgf("Found children, unassigning namespaces from master")
			zkm.unassignLeaderNamespaces()
		}
		// Setup watcher on namespace
		go zkm.watchNamespace()
		// Watch for new nodes
		go zkm.watchNodes()
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// Setup watcher on namespace
		go zkm.watchNamespace()
		// Setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		go zkm.watchNextNode(watchChan)

	}
}

func (zkm *ZkManager) unassignLeaderNamespaces() {
	// Read map to get all our namespaces
	data, stat, err := zkm.c.Get("/nodeToNamespaceMap")
	if err != nil {
		panic(err)
	}
	nsmd := NodeToNamespaceMapData{}
	err = json.Unmarshal(data, &nsmd)
	if err != nil {
		panic(err)
	}
	namespaces := nsmd.Map[zkm.nodeName]
	logger.Info().Msgf("Became leader, unnassigning namespaces %v", namespaces)

	// Remove namespaces from zk
	for ns := range namespaces {
		_, nodeStat, err := zkm.c.Get("/namespaceToNode")
		if err != nil {
			panic(err)
		}
		err = zkm.c.Delete("/namespaceToNode/"+ns, nodeStat.Version)
		if err != nil {
			panic(err)
		}
	}

	// Modify and write back map
	delete(nsmd.Map, zkm.nodeName)
	data, err = json.Marshal(nsmd)
	if err != nil {
		panic(err)
	}
	_, err = zkm.c.Set("/nodeToNamespaceMap", data, stat.Version)
	if err != nil {
		panic(err)
	}

	// Deactivate namespaces
	for ns := range namespaces {
		zkm.nsm.DeactivateNamespace(ns)
	}
}

func (zkm *ZkManager) watchNodes() {
	children, _, nodesChan, err := zkm.c.ChildrenW("/nodes")
	if err != nil {
		panic(err)
	}
	zkm.watchNodesChan = nodesChan

	for {
		childrenMap := map[string]bool{}
		for _, c := range children {
			childrenMap[c] = true
		}
		zkm.nodeMap = childrenMap
		zkm.DistributeNamespaces()
		e := <-zkm.watchNodesChan
		if !zkm.isLeader {
			logger.Info().Msgf("Detected change in agg nodes, but no longer leader")
			return
		}
		logger.Info().Msgf("Detected change in agg nodes")
		if e.Type != zk.EventNodeChildrenChanged {
			panic(fmt.Sprintf("ZK - Unexpected event in watchNodes -  %s (%d)", e.Type.String(), e.Type))
		}
		children, _, nodesChan, err = zkm.c.ChildrenW("/nodes")
		if err != nil {
			panic(err)
		}
		zkm.watchNodesChan = nodesChan
	}
}

func (zkm *ZkManager) watchNamespace() {
	// TODO Change map to be per node. Have leader watch on parent dir while children watch only their own path
	for {
		data, _, nsmChan, err := zkm.c.GetW("/nodeToNamespaceMap")
		if err != nil {
			panic(err)
		}
		nsmd := NodeToNamespaceMapData{}
		err = json.Unmarshal(data, &nsmd)
		if err != nil {
			panic(err)
		}
		// Get our current namespaces and compare with new ones to find ones to deactivate
		// TODO Deep copy
		nsToDeactivate := map[string]bool{}
		for ns := range zkm.nsm.ActiveNamespaces {
			nsToDeactivate[ns] = true
		}
		for ns := range nsmd.Map[zkm.nodeName] {
			zkm.nsm.ActivateNamespace(ns)
			delete(nsToDeactivate, ns)

		}
		for ns := range nsToDeactivate {
			logger.Info().Msgf("Deactivating namespace %s", ns)
			zkm.nsm.DeactivateNamespace(ns)
		}
		logger.Info().Msgf("Updated namespace map %+v", zkm.nsm.ActiveNamespaces)
		e := <-nsmChan
		if e.Type != zk.EventNodeDataChanged {
			panic(fmt.Sprintf("ZK - Unexpected event in watchNamespace - %s (%d)", e.Type.String(), e.Type))
		}
	}
}

func (zkm *ZkManager) getWatchNodeChannel(votes []string) <-chan zk.Event {
	watchVote := ""
	for i, v := range votes {
		if v == zkm.currVote {
			if i != 0 {
				watchVote = votes[i-1]
			} else {
				panic("Something went wrong with zk")
			}
		}
	}
	exists, _, watchEvent, err := zkm.c.ExistsW("/election/" + watchVote)
	if !exists {
		panic("ZK - node disappeared")
	}
	if err != nil {
		panic(err)
	}
	logger.Info().Msgf("Watching vote %s", watchVote)
	return watchEvent
}

func (zkm *ZkManager) watchNextNode(ch <-chan zk.Event) {
	e := <-ch
	if e.Type != zk.EventNodeDeleted {
		panic("Zk - Invalid operation on vote")
	}
	logger.Info().Msgf("Detected node status change")
	// Check votes
	votes, _, err := zkm.c.Children("/election")
	if err != nil {
		panic(err)
	}
	sort.Strings(votes)

	if votes[0] == zkm.currVote {
		// If smallest vote, become leader
		logger.Info().Msgf("Became leader")
		zkm.isLeader = true
		// Unassign the nodes owned by this node, since it became leader
		// They will be redistributed in watchNodes, only if there are other
		// nodes
		if len(zkm.nodeMap) > 1 {
			logger.Info().Msgf("Found children, unassigning namespaces from master")
			zkm.unassignLeaderNamespaces()
		}
		// Watch for new nodes
		zkm.watchNodes()
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// If not leader, setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		zkm.watchNextNode(watchChan)
	}
}

func (zkm *ZkManager) IngestConfigToZK(data []byte) {
	// Extract namespace
	var doc map[string]interface{}
	err := json.Unmarshal(data, &doc)
	if err != nil {
		panic(err)
	}
	ns := doc["namespace"].(string)
	logger.Info().Msgf("Ingesting config for ns %v into ZK", ns)
	// If exists, set, otherwise, create
	_, stat, err := zkm.c.Get("/configs/" + ns)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			_, err := zkm.c.Create("/configs/"+ns, data, 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		_, err := zkm.c.Set("/configs/"+ns, data, stat.Version)
		if err != nil {
			panic(err)
		}
	}
}

func (zkm *ZkManager) ingestConfigsToZK(configFiles []string) {
	for _, c := range configFiles {
		data, err := ioutil.ReadFile(c)
		if err != nil {
			log.Fatal().Err(err)
		}
		zkm.IngestConfigToZK(data)
	}
}

func mergeChans(signal <-chan bool, in <-chan zk.Event, out chan<- zk.Event) {
	for {
		select {
		case e := <-in:
			out <- e
		case <-signal:
			return
		}
	}
}

func (zkm *ZkManager) watchConfigs() {
	signalChan := make(chan bool)
	e := zk.Event{}
	for {
		// Get our existing configs to find configs to delete
		configsToDelete := map[string]bool{}
		for ns := range zkm.nsm.EventConfigsByNamespace {
			configsToDelete[ns] = true
		}
		// Create new chans
		configs, _, parentChan, err := zkm.c.ChildrenW("/configs")
		if err != nil {
			panic(err)
		}
		watchChan := make(chan zk.Event)
		go mergeChans(signalChan, parentChan, watchChan)
		logger.Info().Msgf("Setting configs %+v", configs)
		for _, ns := range configs {
			data, _, nodeChan, err := zkm.c.GetW("/configs/" + ns)
			if err != nil {
				panic(err)
			}
			zkm.nsm.SetNamespaceFromData(data)
			go mergeChans(signalChan, nodeChan, watchChan)
			delete(configsToDelete, ns)
		}
		// Find namespaces that were deleted and deactivate them
		for ns := range configsToDelete {
			delete(zkm.nsm.MetricConfigsByNamespace, ns)
			delete(zkm.nsm.NsMetadata, ns)
			delete(zkm.nsm.EventConfigsByNamespace, ns)
			delete(zkm.nsm.MetricConfigsByNamespace, ns)
			delete(zkm.nsm.EventToMetricMap, ns)
		}

		// If new namespaces were added, we need to distribute them
		if zkm.isLeader && e.Type == zk.EventNodeChildrenChanged {
			zkm.DistributeNamespaces()
		}
		e = <-watchChan
		logger.Info().Msgf("Config change: %s", e.Type.String())
		// kill all the other channels waiting
		if len(configs) > 1 {
			signalChan <- true
		}
	}
}

func (zkm *ZkManager) metadataUpdater() {
	for {
		ns := <-zkm.updateMetadataChan
		nsMeta := zkm.nsm.NsMetadata[ns]
		data, err := json.Marshal(nsMeta)
		if err != nil {
			panic(err)
		}
		_, stat, err := zkm.c.Get("/namespaceMetadata/" + ns)
		if err != nil {
			if !errors.Is(err, zk.ErrNoNode) {
				panic(err)
			}
			_, err = zkm.c.Create("/namespaceMetadata/"+ns, data, 0, zk.WorldACL(zk.PermAll))
			if err != nil {
				panic(err)
			}
			continue
		}
		_, err = zkm.c.Set("/namespaceMetadata/"+ns, data, stat.Version)
		if err != nil {
			panic(err)
		}
	}
}

func (zkm *ZkManager) watchMetadata() {
	signalChan := make(chan bool)
	e := zk.Event{}
	for {
		// Compare against existing metadata to find namespaces to delete
		nsToDelete := map[string]bool{}
		for ns := range zkm.nsm.NsMetadata {
			nsToDelete[ns] = true
		}
		// Get ZK metadata
		metadata, _, parentChan, err := zkm.c.ChildrenW("/namespaceMetadata")
		if err != nil {
			panic(err)
		}
		watchChan := make(chan zk.Event)
		go mergeChans(signalChan, parentChan, watchChan)
		logger.Info().Msgf("Setting metadata %+v", metadata)
		for _, ns := range metadata {
			data, _, nodeChan, err := zkm.c.GetW("/namespaceMetadata/" + ns)
			if err != nil {
				panic(err)
			}
			nsMeta := agg.NamespaceMetadata{}
			err = json.Unmarshal(data, &nsMeta)
			if err != nil {
				panic(err)
			}
			zkm.nsm.NsMetadata[ns] = nsMeta
			go mergeChans(signalChan, nodeChan, watchChan)
			delete(nsToDelete, ns)
		}
		e = <-watchChan
		logger.Info().Msgf("Metadata change: %s", e.Type.String())
		// Kill all the other channels waiting
		if len(metadata) > 0 {
			signalChan <- true
		}
		// Cleanup deleted metadata
		for ns := range nsToDelete {
			delete(zkm.nsm.NsMetadata, ns)
		}
	}
}

func (zkm *ZkManager) DeleteNamespace(ns string) error {
	// Get name of node from /namespaceToNode
	data, stat, err := zkm.c.Get("/namespaceToNode/" + ns)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return &agg.NamespaceNotFoundError{}
		} else {
			panic(err)
		}
	}
	nsToNode := NamespaceToNodeData{}
	err = json.Unmarshal(data, &nsToNode)
	if err != nil {
		panic(err)
	}
	// Remove from /namespaceToNode
	err = zkm.c.Delete("/namespaceToNode/"+ns, stat.Version)
	if err != nil {
		panic(err)
	}
	// Revove from /namespaceMetadata (causes metadata deletion on node)
	_, stat, err = zkm.c.Get("/namespaceMetadata/" + ns)
	if err != nil {
		panic(err)
	}
	err = zkm.c.Delete("/namespaceMetadata/"+ns, stat.Version)
	if err != nil {
		panic(err)
	}
	// Remove from /nodeToNamespaceMap (causes namespace deactivation on node)
	data, stat, err = zkm.c.Get("/nodeToNamespaceMap")
	if err != nil {
		panic(err)
	}
	nsmd := NodeToNamespaceMapData{}
	err = json.Unmarshal(data, &nsmd)
	if err != nil {
		panic(err)
	}
	delete(nsmd.Map[nsToNode.Node], ns)
	data, err = json.Marshal(nsmd)
	if err != nil {
		panic(err)
	}
	_, err = zkm.c.Set("/nodeToNamespaceMap", data, stat.Version)
	if err != nil {
		panic(err)
	}
	// Remove from /configs (causes config data deletion from all nodes)
	_, stat, err = zkm.c.Get("/configs/" + ns)
	if err != nil {
		panic(err)
	}
	err = zkm.c.Delete("/configs/"+ns, stat.Version)
	if err != nil {
		panic(err)
	}
	return nil
}

func (zkm *ZkManager) GetConfig(ns string) (string, error) {
	data, _, err := zkm.c.Get("/configs/" + ns)
	if err != nil {
		if errors.Is(err, zk.ErrNoNode) {
			return "", &agg.NamespaceNotFoundError{}
		} else {
			panic(err)
		}
	}
	cfg := map[string]interface{}{}
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}
	data, err = json.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return string(data), nil
}
