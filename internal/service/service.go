package service

import (
	"io/ioutil"
	"log"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

// Service contains the complete running aggregator service
type Service struct {
	e agg.Engine
}

// MakeNewService creates and initializes the aggregator service
func MakeNewService(rocksDBPath string) Service {
	storage := agg.NewRocksDBStorage(rocksDBPath)
	parser := agg.NSMFromRaw(getConfigText(), storage)
	engine := agg.NewEngine(&parser)
	svc := Service{e: engine}
	return svc
}

func getConfigText() []byte {
	content, err := ioutil.ReadFile("config/example")
	if err != nil {
		log.Fatal(err)
	}
	return content
}
