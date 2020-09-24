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
)

// ZkManager handles all zookeeper interaction
type ZkManager struct {
	c             *zk.Conn
	localOnlyMode bool
	nodeName      string
	currVote      string
	isLeader      bool
	logger        zerolog.Logger
	watchNodeChan <-chan zk.Event
}

type AggNodeStatus struct {
	isReady bool
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
func MakeNewZkManager(zkURL string, nodeName string) *ZkManager {
	logger := zerolog.New(os.Stderr).With().Str("source", "ZK").Logger()

	if zkURL == "" {
		logger.Info().Msgf("Local only mode")
		return &ZkManager{
			c:             nil,
			localOnlyMode: true,
			nodeName:      nodeName,
			isLeader:      false,
			logger:        logger,
			watchNodeChan: nil,
		}
	}
	l := ZkLogger{}
	opt := zk.WithLogger(l)
	c, _, err := zk.Connect([]string{zkURL}, time.Second, opt)

	if err != nil {
		panic(err)
	}

	zkm := &ZkManager{
		c:             c,
		localOnlyMode: false,
		nodeName:      nodeName,
		isLeader:      false,
		logger:        logger,
		watchNodeChan: nil,
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
		isReady: true,
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
		panic(err)
	}
	sort.Strings(votes)
	if votes[0] == zkm.currVote {
		// If smallest vote, become leader
		logger.Info().Msgf("Became leader")
		zkm.isLeader = true
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// TODO
		// If not leader, setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		zkm.watchNextNode(watchChan)
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
	if zkm.watchNodeChan != nil {
		panic("Zk - Already watching")
	}
	zkm.watchNodeChan = ch
	e := <-zkm.watchNodeChan
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
		zkm.watchNodeChan = nil
	} else {
		logger.Info().Msgf("Not leader, setting up leader watch")
		// If not leader, setup watcher on next smallest node
		watchChan := zkm.getWatchNodeChannel(votes)
		zkm.watchNodeChan = nil
		zkm.watchNextNode(watchChan)
	}
}

// TODO Init nodes to namespace

// TODO Init namespace to node

// TODO Register for changes in namespace map

// TODO Master: Distribute namespaces
