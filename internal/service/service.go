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
func MakeNewService(redisURL string, zkURL string, nodeName string) Service {
	storage := agg.NewRedisStorage(redisURL)
	nsm := agg.NSMFromRaw(getConfigText(), storage)
	engine := agg.NewEngine(&nsm)
	zkm := zk.MakeNewZkManager(zkURL, nodeName, &nsm)
	svc := Service{
		e:        engine,
		zkm:      zkm,
		nodeName: nodeName,
	}
	return svc
}

func getConfigText() []byte {
	content, err := ioutil.ReadFile("config/example")
	if err != nil {
		log.Fatal(err)
	}
	return content
}
