package zkManager

import (
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/rs/zerolog/log"
)

type ZkManager struct {
	c             *zk.Conn
	localOnlyMode bool
}

func MakeNewZkManager(zkURL string) *ZkManager {
	if zkURL == "" {
		log.Info().Msgf("Local only mode")
		return &ZkManager{
			c:             nil,
			localOnlyMode: true,
		}
	}
	c, _, err := zk.Connect([]string{zkURL}, time.Second)
	if err != nil {
		panic(err)
	}

	return &ZkManager{
		c:             c,
		localOnlyMode: false,
	}
}
