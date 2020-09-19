package zkManager

import (
	"fmt"
	"time"

	"github.com/go-zookeeper/zk"
)

type ZkManager struct {
	c             *zk.Conn
	localOnlyMode bool
}

func MakeNewZkManager(zkURL string) *ZkManager {
	if zkURL == "" {
		fmt.Printf("Starting in local only mode\n")
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
