package service

import (
	agg "github.com/kal-g/aggregator-go/internal/aggregator"
	"github.com/kal-g/aggregator-go/internal/zk"
)

// Service contains the complete running aggregator service
type Service struct {
	e        agg.Engine
	zkm      *zk.ZkManager
	nodeName string
	Nsm      *agg.NamespaceManager
}

// MakeNewService creates and initializes the aggregator service
func MakeNewService(redisURL string, zkURL string, nodeName string) Service {
	storage := agg.NewRedisStorage(redisURL)
	nsm := agg.NewNSM(storage, zkURL == "")
	// Get list of configs
	engine := agg.NewEngine(&nsm)
	zkm := zk.MakeNewZkManager(zkURL, nodeName, &nsm)
	svc := Service{
		e:        engine,
		zkm:      zkm,
		nodeName: nodeName,
		Nsm:      &nsm,
	}
	return svc
}
