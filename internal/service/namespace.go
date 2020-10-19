package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

type NamespaceGetInfoResult struct {
	Err  error                 `json:"error"`
	Data agg.NamespaceMetadata `json:"data"`
}

func (s *Service) NamespaceGetInfo(w http.ResponseWriter, r *http.Request) {
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

	// Check if namespace exists
	s.e.Nsm.NsDataLck.RLock()
	meta, exists := s.e.Nsm.ActiveNamespaces[namespace]
	s.e.Nsm.NsDataLck.RUnlock()

	res := NamespaceGetInfoResult{}
	if !exists {
		res.Err = &agg.NamespaceNotFoundError{}
	} else {
		res.Data = meta
	}
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
