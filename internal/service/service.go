package service

import (
	"os"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
	"github.com/kal-g/aggregator-go/internal/zk"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger = zerolog.New(os.Stderr).With().
	Str("source", "SVC").
	Timestamp().
	Logger()

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
	return svc
}
