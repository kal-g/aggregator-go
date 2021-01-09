package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

type NamespaceGetInfoResult struct {
	Err  string                `json:"error"`
	Data agg.NamespaceMetadata `json:"data"`
}

type NamespaceSetResult struct {
	Err string `json:"error"`
}

type NamespaceDeleteResult struct {
	Err string `json:"error"`
}

type NamespaceGetResult struct {
	Cfg string `json:"cfg"`
	Err string `json:"error"`
}

func (s *Service) NamespaceGetConfig(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		panic(err)
	}

	n, namespaceSet := bodyJSON["namespace"]
	if !namespaceSet {
		panic(err)
	}
	ns, isString := n.(string)
	if !isString {
		// TODO error
		panic("Type error in get namespace")
	}

	cfg, err := s.Zkm.GetConfig(ns)
	res := NamespaceGetResult{}
	if err != nil {
		res.Err = err.Error()
	} else {
		res.Cfg = cfg
	}
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Service) NamespaceSetConfig(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		panic(err)
	}

	overwriteIfExists := false
	namespaceConfig := map[string]interface{}{}

	if oie, oieSet := bodyJSON["overwriteIfExists"]; oieSet {
		oieBool, isBool := oie.(bool)
		if !isBool {
			// TODO Error
			panic("Type error in set namespace oie")
		}
		overwriteIfExists = oieBool
	}

	if nsCfg, nsCfgSet := bodyJSON["namespaceConfig"]; nsCfgSet {
		nsCfgJSON, isJSON := nsCfg.(map[string]interface{})
		if !isJSON {
			// TODO Error
			panic("Type error in set namespace nsCfg")
		}
		namespaceConfig = nsCfgJSON
	}

	res := NamespaceSetResult{}
	// Extract namespace
	ns := namespaceConfig["namespace"].(string)

	if !overwriteIfExists {
		if _, exists := s.e.Nsm.EventConfigsByNamespace[ns]; exists {
			err := &agg.NamespaceExistsError{}
			res.Err = err.Error()
			data, _ := json.Marshal(res)
			w.Header().Set("Content-Type", "application/json")
			_, werr := w.Write(data)
			if werr != nil {
				panic(err)
			}
			return
		}
	}
	// Add JSON
	cfgData, _ := json.Marshal(namespaceConfig)
	logger.Info().Msgf("Ingesting config %s", ns)
	s.Zkm.IngestConfigToZK(cfgData)
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Service) NamespaceDelete(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		panic(err)
	}

	n, namespaceSet := bodyJSON["namespace"]
	if !namespaceSet {
		panic(err)
	}
	ns, isString := n.(string)
	if !isString {
		// TODO error
		panic("Type error in get namespace")
	}
	err = s.Zkm.DeleteNamespace(ns)

	res := NamespaceDeleteResult{}
	if err != nil {
		res.Err = err.Error()
	}
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Service) NamespaceGetInfo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	bodyJSON := map[string]interface{}{}
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		panic(err)
	}

	namespace := ""
	if n, namespaceSet := bodyJSON["namespace"]; namespaceSet {
		nString, isString := n.(string)
		if !isString {
			// TODO error
			panic("Type error in get namespace")
		}
		namespace = nString
	}

	// Check if namespace exists
	s.e.Nsm.NsDataLck.RLock()
	meta, exists := s.e.Nsm.NsMetadata[namespace]
	s.e.Nsm.NsDataLck.RUnlock()

	res := NamespaceGetInfoResult{}
	if !exists {
		err := &agg.NamespaceNotFoundError{}
		res.Err = err.Error()
	} else {
		res.Data = meta
	}
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}
