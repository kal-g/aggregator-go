package aggregator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// TODO Make fully configurable

type Service struct {
	e Engine
}

type ConsumeResult struct {
	ErrorCode EngineHandleResult `json:"error_code"`
}

func MakeNewService(rocksDBPath string) Service {
	storage := NewRocksDBStorage(rocksDBPath)
	parser := NewConfigParserFromRaw(getConfigText(), storage)
	engine := NewEngine(&parser)
	svc := Service{e: engine}
	return svc
}

// Consume is the endpoint that ingests event into aggregator
func (s *Service) Consume(w http.ResponseWriter, r *http.Request) {
	isVerbose := false
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	re := map[string]interface{}{}
	json.Unmarshal(body, &re)
	// Check options
	if _, verbose := re["verbose"]; verbose {
		isVerbose = true
	}
	payload := re["payload"].(map[string]interface{})

	// Since we're unmarshalling into an interface, unmarshal converts to floats
	// Convert the floats to ints
	sanitizedPayload := map[string]interface{}{}
	for k, v := range payload {
		vAsFloat, isFloat := v.(float64)
		if isFloat {
			sanitizedPayload[k] = int(vAsFloat)
		} else {
			sanitizedPayload[k] = v
		}
	}

	engineResult := s.e.HandleRawEvent(sanitizedPayload)

	if isVerbose {
		consumeRes := ConsumeResult{
			ErrorCode: engineResult,
		}
		data, _ := json.Marshal(consumeRes)
		t := ConsumeResult{}
		json.Unmarshal(data, &t)

		fmt.Printf("%+v %+v %+v\n", consumeRes, data, t)

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func getConfigText() []byte {
	content, err := ioutil.ReadFile("tools/config/example")
	if err != nil {
		log.Fatal(err)
	}
	return content
}
