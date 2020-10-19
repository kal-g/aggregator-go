package service

import (
	"io/ioutil"
	"log"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
	"github.com/kal-g/aggregator-go/internal/zk"
)

// Service contains the complete running aggregator service
type Service struct {
	e        agg.Engine
	zkm      *zk.ZkManager
	nodeName string
}

// MakeNewService creates and initializes the aggregator service
func MakeNewService(redisURL string, zkURL string, nodeName string, configFile string) Service {
	storage := agg.NewRedisStorage(redisURL)
	config := []byte{}
	if configFile != "" {
		parsedConfig, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
		config = parsedConfig
	}

	nsm := agg.NSMFromRaw(config, storage, zkURL == "")
	engine := agg.NewEngine(&nsm)
	zkm := zk.MakeNewZkManager(zkURL, nodeName, &nsm)
	svc := Service{
		e:        engine,
		zkm:      zkm,
		nodeName: nodeName,
	}
	return svc
}
