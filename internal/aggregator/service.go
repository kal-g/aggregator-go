package aggregator

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

// TODO Make fully configurable

// Service contains the complete running aggregator service
type Service struct {
	e engine
}

// ConsumeResult contains the result of consumption and any error codes
type ConsumeResult struct {
	ErrorCode engineHandleResult `json:"error_code"`
}

// MakeNewService creates and initializes the aggregator service
func MakeNewService(rocksDBPath string) Service {
	storage := newRocksDBStorage(rocksDBPath)
	parser := newConfigParserFromRaw(getConfigText(), storage)
	engine := newEngine(&parser)
	svc := Service{e: engine}
	return svc
}

// Consume is the endpoint that ingests event into aggregator
func (s *Service) Consume(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	re := map[string]interface{}{}
	json.Unmarshal(body, &re)

	// Check options
	isVerbose := false
	namespace := ""
	if _, verbose := re["verbose"]; verbose {
		isVerbose = true
	}
	if n, namespaceSet := re["namespace"]; namespaceSet {
		nString, isString := n.(string)
		if !isString {
			// TODO Return error
		}
		namespace = nString
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

	engineResult := s.e.HandleRawEvent(sanitizedPayload, namespace)

	if isVerbose {
		consumeRes := ConsumeResult{
			ErrorCode: engineResult,
		}
		data, _ := json.Marshal(consumeRes)
		t := ConsumeResult{}
		json.Unmarshal(data, &t)

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func getConfigText() []byte {
	content, err := ioutil.ReadFile("config/example")
	if err != nil {
		log.Fatal(err)
	}
	return content
}
