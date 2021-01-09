package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	agg "github.com/kal-g/aggregator-go/internal/aggregator"
)

type CountResult struct {
	Count int
}

// Count is the endpoint that returns the count of a particular metric
func (s *Service) Count(w http.ResponseWriter, r *http.Request) {
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
		nString, _ := n.(string)
		/*if !isString {
			// TODO Return error
		}*/
		namespace = nString
	}

	metricKey, metricKeyExists := bodyJSON["metricKey"]
	metricID, metricIDExists := bodyJSON["metricID"]

	err = nil
	if !metricKeyExists {
		err = &agg.MetricKeyNotFoundError{}
	}
	if !metricIDExists {
		err = &agg.MetricIDNotFoundError{}
	}

	if err != nil {
		res := agg.MetricCountResult{
			Err: err,
		}
		data, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
		return
	}

	metricKeyAsFloat, keyIsFloat := metricKey.(float64)
	metricIDAsFloat, idIsFloat := metricID.(float64)

	if !keyIsFloat {
		err = &agg.MetricKeyInvalidType{}
	}
	if !idIsFloat {
		err = &agg.MetricIDInvalidType{}
	}

	if err != nil {
		res := agg.MetricCountResult{
			Err: err,
		}
		data, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(data)
		if err != nil {
			panic(err)
		}
		return
	}

	countRes := s.doCount(int(metricKeyAsFloat), int(metricIDAsFloat), namespace)

	data, _ := json.Marshal(countRes)

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}

func (s *Service) doCount(metricKey int, metricID int, namespace string) agg.MetricCountResult {
	return s.e.GetMetricCount(namespace, metricKey, metricID)
}
