package zk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/rs/zerolog"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

// ZkManager handles all zookeeper interaction
type ZkManager struct {
	c               *zk.Conn
	localOnlyMode   bool
	nodeName        string
	currVote        string
	isLeader        bool
	logger          zerolog.Logger
	watchLeaderChan <-chan zk.Event
	nsmChan         <-chan zk.Event
	watchNodesChan  <-chan zk.Event
	nsm             *agg.NamespaceManager
}

type AggNodeStatus struct {
	IsReady bool
}

type NamespaceMapData struct {
	NamespaceMap map[string]bool
}

type NamespaceToNodeData struct {
	Node string
}

var logger zerolog.Logger = zerolog.New(os.Stderr).With().Str("source", "ZK").Logger()

// ZkLogger implements logging interface for go-zookeeper package
type ZkLogger struct{}

// Printf implement logger interface
func (l ZkLogger) Printf(fmt string, args ...interface{}) {
	logger.Info().Msgf(fmt, args...)
}

// MakeNewZkManager inits and connects to zk
// If no url given, sets local only mode
func MakeNewZkManager(zkURL string, nodeName string, nsm *agg.NamespaceManager) *ZkManager {
	logger := zerolog.New(os.Stderr).With().Str("source", "ZK").Logger()

	if zkURL == "" {
		logger.Info().Msgf("Local only mode")
		return &ZkManager{
			c:               nil,
			localOnlyMode:   true,
			nodeName:        nodeName,
			isLeader:        false,
			logger:          logger,
			watchLeaderChan: nil,
			nsmChan:         nil,
			watchNodesChan:  nil,
			nsm:             nsm,
		}
	}
	l := ZkLogger{}
	opt := zk.WithLogger(l)
	c, _, err := zk.Connect([]string{zkURL}, time.Second, opt)

	if err != nil {
		panic(err)
	}

	zkm := &ZkManager{
		c:               c,
		localOnlyMode:   false,
		nodeName:        nodeName,
		isLeader:        false,
		logger:          logger,
		watchLeaderChan: nil,
		nsmChan:         nil,
		watchNodesChan:  nil,
		nsm:             nsm,
	}
	zkm.Setup()
	go zkm.LeaderElection()
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
	nodeMap := NamespaceMapData{
		NamespaceMap: map[string]bool{},
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

}

func (zkm *ZkManager) DistributeNamespaces(children []string) {
	if len(children) == 1 {
		logger.Info().Msgf("Only master, no namespaces distributed")
		return
	}
	// TODO Get distributed namespaces
	distributedNs := map[string]bool{}
	nss, _, err := zkm.c.Children("/namespaceToNode")
	if err != nil {
		panic(err)
	}

	for _, ns := range nss {
		distributedNs[ns] = true
	}

	// TODO check against metric map to find non distributed namespaces
	nonDistributedNs := []string{}
	for ns := range zkm.nsm.MetricMap {
		if _, exists := distributedNs[ns]; !exists {
			nonDistributedNs = append(nonDistributedNs, ns)
		}
	}
	logger.Info().Msgf("Namespaces about to be distributed: %+v", nonDistributedNs)

	// Find first non master node
	for _, c := range children {
		if zkm.nodeName != c {
			// Put all non distributed namespaces into map for first non master node
			data, _, err := zkm.c.Get("/nodeToNamespaceMap")
			nsmap := NamespaceMapData{}
			err = json.Unmarshal(data, &nsmap)
			if err != nil {
				logger.Error().Msgf("Error getting nodeToNamespaceMap")
				panic(err)
			}
			for _, ns := range nonDistributedNs {
				nsmap.NamespaceMap[ns] = true
			}
			// Create entries for all new namespaceToNode
			for _, ns := range nonDistributedNs {
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
			break
		}
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
		zkm.watchLeaderChan = nil
		// Watch for new nodes
		zkm.watchNodes()
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// Setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		go zkm.watchNextNode(watchChan)
		// Read and setup watcher on nodeToNamespaceMap
		logger.Info().Msgf("Reading initial namespace map")
		data, _, nsmChan, err := zkm.c.GetW("/nodeToNamespaceMap/" + zkm.nodeName)
		if !errors.As(err, &zk.ErrNoNode) {
			logger.Error().Msgf("Error getting initial namespace map")
			panic(err)
		}

		nsmd := NamespaceMapData{}
		json.Unmarshal(data, &nsmd)
		for ns, _ := range nsmd.NamespaceMap {
			zkm.nsm.ActivateNamespace(ns)
		}
		logger.Info().Msgf("Updated namespace map %+v", zkm.nsm.NsMetaMap)
		zkm.nsmChan = nsmChan
		go zkm.watchNamespace()
	}
}

func (zkm *ZkManager) watchNodes() {
	children, _, nodesChan, err := zkm.c.ChildrenW("/nodes")
	if err != nil {
		panic(err)
	}
	zkm.watchNodesChan = nodesChan
	for {
		zkm.DistributeNamespaces(children)
		e := <-zkm.watchNodesChan
		if e.Type != zk.EventNodeChildrenChanged {
			panic("ZK - Unexpected event")
		}
		children, _, nodesChan, err = zkm.c.ChildrenW("/nodes")
		if err != nil {
			panic(err)
		}
		zkm.watchNodesChan = nodesChan
	}
}

func (zkm *ZkManager) watchNamespace() {
	for {
		e := <-zkm.nsmChan
		if e.Type != zk.EventNodeDataChanged {
			panic("ZK - Unexpected event")
		}
		data, _, err := zkm.c.Get("/nodeToNamespaceMap/" + zkm.nodeName)
		if !errors.As(err, &zk.ErrNoNode) {
			panic(err)
		}
		nsmd := NamespaceMapData{}
		json.Unmarshal(data, &nsmd)
		for ns, _ := range nsmd.NamespaceMap {
			zkm.nsm.ActivateNamespace(ns)
		}
		logger.Info().Msgf("Updated namespace map %+v", zkm.nsm.NsMetaMap)
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
	if zkm.watchLeaderChan != nil {
		panic("Zk - Already watching next node")
	}
	zkm.watchLeaderChan = ch
	e := <-zkm.watchLeaderChan
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
		zkm.watchLeaderChan = nil
		// Watch for new nodes
		zkm.watchNodes()
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// If not leader, setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		zkm.watchLeaderChan = nil
		zkm.watchNextNode(watchChan)
	}
}

// TODO Init nodes to namespace

// TODO Init namespace to node

// TODO Register for changes in namespace map

// TODO Master: Distribute namespaces
