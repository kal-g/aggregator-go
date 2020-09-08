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
		res := agg.MetricCountResult{
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
		res := agg.MetricCountResult{
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

func (s *Service) doCount(metricKey int, metricID int, namespace string) agg.MetricCountResult {
	return s.e.GetMetricCount(namespace, metricKey, metricID)
}
