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

type CountResult struct {
	Count int
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
	bodyJSON := map[string]interface{}{}
	json.Unmarshal(body, &bodyJSON)

	// Check options
	isVerbose := false
	namespace := ""
	if _, verbose := bodyJSON["verbose"]; verbose {
		isVerbose = true
	}
	if n, namespaceSet := bodyJSON["namespace"]; namespaceSet {
		nString, isString := n.(string)
		if !isString {
			// TODO Return error
		}
		namespace = nString
	}

	engineResult := deferredSuccess
	if isVerbose {
		engineResult = s.doConsume(bodyJSON["payload"].(map[string]interface{}), namespace)
	} else {
		go s.doConsume(bodyJSON["payload"].(map[string]interface{}), namespace)

	}

	consumeRes := ConsumeResult{
		ErrorCode: engineResult,
	}
	data, _ := json.Marshal(consumeRes)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// Count is the endpoint that returns the count of a particular metric
func (s *Service) Count(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	json.Unmarshal(body, &bodyJSON)

	namespace := ""
	if n, namespaceSet := bodyJSON["namespace"]; namespaceSet {
		nString, isString := n.(string)
		if !isString {
			// TODO Return error
		}
		namespace = nString
	}

	metricKey, metricKeyExists := bodyJSON["metricKey"]
	metricID, metricIDExists := bodyJSON["metricID"]

	errCode := 0
	if !metricKeyExists {
		errCode = 1
	}
	if !metricIDExists {
		errCode = 2
	}

	if errCode != 0 {
		res := metricCountResult{
			ErrCode: errCode,
		}
		data, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	metricKeyAsFloat, keyIsFloat := metricKey.(float64)
	metricIDAsFloat, idIsFloat := metricID.(float64)

	if !keyIsFloat {
		errCode = 3
	}
	if !idIsFloat {
		errCode = 4
	}

	if errCode != 0 {
		res := metricCountResult{
			ErrCode: errCode,
		}
		data, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}

	countRes := s.doCount(int(metricKeyAsFloat), int(metricIDAsFloat), namespace)

	data, _ := json.Marshal(countRes)

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Service) doCount(metricKey int, metricID int, namespace string) metricCountResult {
	mc := s.e.getMetricConfig(namespace, metricID)
	if mc == nil {
		return metricCountResult{
			ErrCode: 1,
			Count:   0,
		}
	}
	return mc.getCount(metricKey)
}

func (s *Service) doConsume(payload map[string]interface{}, namespace string) engineHandleResult {
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
	return s.e.HandleRawEvent(sanitizedPayload, namespace)
}

func getConfigText() []byte {
	content, err := ioutil.ReadFile("config/example")
	if err != nil {
		log.Fatal(err)
	}
	return content
}
