package zkmanager

import (
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
	}
	zkm.Setup()
	zkm.LeaderElection()
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
	path := fmt.Sprintf("/nodes/%s", zkm.nodeName)
	_, err = zkm.c.Create(path, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
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
	sort.Strings(votes)
	if votes[0] == zkm.currVote {
		// If smallest vote, become leader
		logger.Info().Msgf("Became leader")
		zkm.isLeader = true
	} else {
		// TODO
		// If not leader, setup watcher on next smallest node
	}
}

// TODO Register for changes in namespace map

// TODO Master: Check connection to ZK

// TODO Master: init namespace directory

// TODO Master: Check namespaces

// TODO Master: Distribute namespaces
