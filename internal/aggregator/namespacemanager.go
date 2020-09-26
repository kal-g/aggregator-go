package aggregator

import (
	"encoding/json"
	"sync"
)

// NamespaceMetadata encapsulates user relevant metadata about the namespace
type NamespaceMetadata struct {
	KeySizeMap map[int]int `json:"metric_keys"`
}

// NamespaceManager manages namespace access and metadata
type NamespaceManager struct {
	EventConfigs     []*eventConfig
	MetricConfigs    []*metricConfig
	storage          AbstractStorage
	NsDataLck        *sync.RWMutex
	EventMap         map[int]*eventConfig
	MetricMap        map[string]map[int]*metricConfig
	EventToMetricMap map[string]map[int][]*metricConfig
	NsMetaMap        map[string]NamespaceMetadata
}

// NSMFromRaw creates a namespace manager from a byte stream
func NSMFromRaw(input []byte, storage AbstractStorage) NamespaceManager {
	var doc map[string]interface{}
	json.Unmarshal(input, &doc)

	nsm := NamespaceManager{
		EventConfigs:  extractEventConfigs(doc),
		MetricConfigs: extractMetricConfigs(doc, storage),
		storage:       storage,
		NsDataLck:     &sync.RWMutex{},
	}
	nsm.initConfigMaps()
	return nsm
}

// NSMFromConfigs creates a namespace manager from configs
func NSMFromConfigs(ecs []*eventConfig, mcs []*metricConfig, storage AbstractStorage) NamespaceManager {
	nsm := NamespaceManager{
		EventConfigs:  ecs,
		MetricConfigs: mcs,
		storage:       storage,
		NsDataLck:     &sync.RWMutex{},
	}
	nsm.initConfigMaps()
	return nsm
}

func (nsm *NamespaceManager) ActivateNamespace(ns string) {
	nsm.NsDataLck.Lock()

	if _, exists := nsm.NsMetaMap[ns]; exists {
		nsm.NsDataLck.Unlock()
		return
	}

	nsm.NsMetaMap[ns] = NamespaceMetadata{
		KeySizeMap: map[int]int{},
	}

	for _, mc := range nsm.MetricConfigs {
		if ns == mc.Namespace {
			// TODO init with old values
			nsm.NsMetaMap[ns].KeySizeMap[mc.ID] = 0
		}

	}
	nsm.NsDataLck.Unlock()
}

func (nsm *NamespaceManager) DeactivateNamespace(ns string) {
	nsm.NsDataLck.Lock()
	delete(nsm.NsMetaMap, ns)
	nsm.NsDataLck.Unlock()
}

// TODO Add zombie state for namespace

func (nsm *NamespaceManager) initConfigMaps() {
	nsMetaMap := make(map[string]NamespaceMetadata)
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
	nsm.NsMetaMap = nsMetaMap

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

func (nsm *NamespaceManager) namespaceRLock(ns string) {
	nsm.NsDataLck.RLock()
}

func (nsm *NamespaceManager) namespaceRUnlock(ns string) {
	nsm.NsDataLck.RUnlock()
}

func (nsm *NamespaceManager) namespaceWLock(ns string) {
	nsm.NsDataLck.Lock()
}

func (nsm *NamespaceManager) namespaceUnlock(ns string) {
	nsm.NsDataLck.Unlock()
}
