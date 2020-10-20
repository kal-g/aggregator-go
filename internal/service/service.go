package service

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"

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
func MakeNewService(redisURL string, zkURL string, nodeName string, configFiles []string) Service {
	storage := agg.NewRedisStorage(redisURL)
	nsm := agg.NewNSM(storage, zkURL == "")
	engine := agg.NewEngine(&nsm)
	zkm := zk.MakeNewZkManager(zkURL, nodeName, &nsm, configFiles)
	svc := Service{
		e:        engine,
		zkm:      zkm,
		nodeName: nodeName,
		Nsm:      &nsm,
	}
	// Read configs files into ZK if provided
	for _, c := range configFiles {
		data, err := ioutil.ReadFile(c)
		if err != nil {
			log.Fatal().Err(err)
		}
		svc.Nsm.SetNamespaceFromData(data)
	}
	return svc
}
