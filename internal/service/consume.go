package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ConsumeResult struct {
	Err string `json:"error"`
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
	namespace := ""
	if n, namespaceSet := bodyJSON["namespace"]; namespaceSet {
		nString, isString := n.(string)
		if !isString {
			// TODO Return error
			panic("Something went wrong")
		}
		namespace = nString
	}
	err = s.doConsume(bodyJSON["payload"].(map[string]interface{}), namespace)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	consumeRes := ConsumeResult{
		Err: errMsg,
	}
	data, _ := json.Marshal(consumeRes)
	logger.Info().Msgf("Send %+v", string(data))
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Service) doConsume(payload map[string]interface{}, namespace string) error {
	// Since we're unmarshalling into an interface, unmarshal converts to floats
	// Convert the floats to ints
	logger.Info().Msgf("DoConsume\n")
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
