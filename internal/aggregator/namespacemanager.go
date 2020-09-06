package aggregator

import (
	"encoding/json"
	"sync"
)

type namespaceManager struct {
	EventConfigs     []*eventConfig
	MetricConfigs    []*metricConfig
	storage          AbstractStorage
	nsLck            map[string]*sync.RWMutex
	metaMtx          sync.Mutex
	EventMap         map[int]*eventConfig
	MetricMap        map[string]map[int]*metricConfig
	EventToMetricMap map[string]map[int][]*metricConfig
}

func newConfigParserFromRaw(input []byte, storage AbstractStorage) namespaceManager {
	var doc map[string]interface{}
	json.Unmarshal(input, &doc)

	nsm := namespaceManager{
		EventConfigs:  extractEventConfigs(doc),
		MetricConfigs: extractMetricConfigs(doc, storage),
		storage:       storage,
		nsLck:         map[string]*sync.RWMutex{},
	}
	nsm.initConfigMaps()
	return nsm
}

func newConfigParserFromConfigs(ecs []*eventConfig, mcs []*metricConfig, storage AbstractStorage) namespaceManager {
	nsm := namespaceManager{
		EventConfigs:  ecs,
		MetricConfigs: mcs,
		storage:       storage,
		nsLck:         map[string]*sync.RWMutex{},
	}
	nsm.initConfigMaps()
	return nsm
}

func (nsm *namespaceManager) initConfigMaps() {
	eventMap := make(map[int]*eventConfig)

	// Create a map from event id to event config
	for _, eventConfig := range nsm.EventConfigs {
		eventMap[eventConfig.ID] = eventConfig
	}

	// Create a map from metric ID to metric
	metricMap := make(map[string]map[int]*metricConfig)
	for _, mc := range nsm.MetricConfigs {
		// Init the namespace if it doesn't exist
		_, namespaceExists := metricMap[mc.Namespace]
		if !namespaceExists {
			metricMap[mc.Namespace] = make(map[int]*metricConfig)
		}
		metricMap[mc.Namespace][mc.ID] = mc
	}

	// Create a map from event id to metric configs
	eventToMetricMap := make(map[string]map[int][]*metricConfig)
	for _, mc := range nsm.MetricConfigs {
		// Init the namespace if it doesn't exist
		_, namespaceExists := eventToMetricMap[mc.Namespace]
		if !namespaceExists {
			eventToMetricMap[mc.Namespace] = make(map[int][]*metricConfig)
		}
		for _, eventID := range mc.EventIds {
			// Initialize the slice if it doesn't exist
			_, metricExists := eventToMetricMap[mc.Namespace][eventID]
			if !metricExists {
				eventToMetricMap[mc.Namespace][eventID] = []*metricConfig{}
			}
			eventToMetricMap[mc.Namespace][eventID] = append(eventToMetricMap[mc.Namespace][eventID], metricMap[mc.Namespace][mc.ID])
		}
	}

	nsm.EventMap = eventMap
	nsm.MetricMap = metricMap
	nsm.EventToMetricMap = eventToMetricMap

}

func extractEventConfigs(doc map[string]interface{}) []*eventConfig {
	reConfigs := doc["events"].([]interface{})
	eConfigs := []*eventConfig{}
	for _, reConfig := range reConfigs {
		reConfig := reConfig.(map[string]interface{})
		// Get the initializers for the event config
		name := reConfig["name"].(string)
		id := int(reConfig["id"].(float64))
		rFields := reConfig["fields"].(map[string]interface{})
		fields := map[string]fieldType{}
		for fName, fType := range rFields {
			fields[fName] = fieldType(fType.(float64))
		}
		// Create event config
		ec := eventConfig{
			Name:   name,
			ID:     id,
			Fields: fields,
		}
		eConfigs = append(eConfigs, &ec)
	}
	return eConfigs
}

func extractMetricConfigs(doc map[string]interface{}, storage AbstractStorage) []*metricConfig {
	rmConfigs := doc["metrics"].([]interface{})
	mConfigs := []*metricConfig{}
	for _, rmConfig := range rmConfigs {
		rmConfig := rmConfig.(map[string]interface{})
		// Get the initializers for the metric config
		name := rmConfig["name"].(string)
		id := int(rmConfig["id"].(float64))
		keyField := rmConfig["key_field"].(string)
		namespace := rmConfig["namespace"].(string)
		// TODO Replace when adding other types
		//metric_type := raw_metric_config["type"].(string)
		countField := rmConfig["count_field"].(string)
		reIDs := rmConfig["event_ids"].([]interface{})
		eIDs := []int{}
		for _, eID := range reIDs {
			eIDs = append(eIDs, int(eID.(float64)))
		}
		// Create metric config
		mc := metricConfig{
			ID:         id,
			Name:       name,
			EventIds:   eIDs,
			KeyField:   keyField,
			CountField: countField,
			MetricType: countMetricType,
			Namespace:  namespace,
			Filter:     extractMetricFilters(rmConfig["filter"].([]interface{})),
			Storage:    storage,
		}
		mConfigs = append(mConfigs, &mc)
	}
	return mConfigs
}

func extractMetricFilters(filt []interface{}) abstractFilter {
	// Name is always the first element
	var f abstractFilter
	filterName := filt[0].(string)
	if filterName == "null" {
		f = NullFilter{}
	} else if filterName == "gt" {
		f = greaterThanFilter{filt[1].(string), int(filt[2].(float64))}
	} else if filterName == "all" {
		filters := []abstractFilter{}
		for i := 1; i < len(filt); i++ {
			filters = append(filters, extractMetricFilters(filt[i].([]interface{})))
		}
		f = allFilter{filters}
	}
	return f
}

func (nsm *namespaceManager) namespaceRLock(ns string) {
	_, exists := nsm.nsLck[ns]
	if !exists {
		// Get global lock, check for existence again
		nsm.metaMtx.Lock()
		_, exists2 := nsm.nsLck[ns]
		if !exists2 {
			// Still doesn't exist, create
			nsm.nsLck[ns] = &sync.RWMutex{}
		}
		nsm.metaMtx.Unlock()
	}
	// Lock must exist at this point
	nsm.nsLck[ns].RLock()
}

func (nsm *namespaceManager) namespaceRUnlock(ns string) {
	nsm.nsLck[ns].RUnlock()
}

func (nsm *namespaceManager) namespaceWLock(ns string) {
	_, exists := nsm.nsLck[ns]
	if !exists {
		// Get global lock, check for existence again
		nsm.metaMtx.Lock()
		_, exists2 := nsm.nsLck[ns]
		if !exists2 {
			// Still doesn't exist, create
			nsm.nsLck[ns] = &sync.RWMutex{}
		}
		nsm.metaMtx.Unlock()
	}
	// Lock must exist at this point
	nsm.nsLck[ns].Lock()
}

func (nsm *namespaceManager) namespaceUnlock(ns string) {
	nsm.nsLck[ns].Unlock()
}