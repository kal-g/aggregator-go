package aggregator

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

// NamespaceMetadata encapsulates user relevant metadata about the namespace
type NamespaceMetadata struct {
	KeySizeMap map[int]int `json:"metric_keys"`
}

// NamespaceManager manages namespace access and metadata
type NamespaceManager struct {
	EventConfigsByNamespace  map[string]map[int]*eventConfig
	MetricConfigsByNamespace map[string]map[int]*metricConfig
	storage                  AbstractStorage
	NsDataLck                *sync.RWMutex
	EventToMetricMap         map[string]map[int][]*metricConfig
	ActiveNamespaces         map[string]NamespaceMetadata
	SingleNodeMode           bool
}

var nsLogger zerolog.Logger = zerolog.New(os.Stderr).With().Str("source", "NSM").Logger()

// NSMFromRaw creates a namespace manager from a byte stream
func NewNSM(storage AbstractStorage, singleNodeMode bool) NamespaceManager {

	nsm := NamespaceManager{
		EventConfigsByNamespace:  map[string]map[int]*eventConfig{},
		MetricConfigsByNamespace: map[string]map[int]*metricConfig{},
		storage:                  storage,
		NsDataLck:                &sync.RWMutex{},
		EventToMetricMap:         map[string]map[int][]*metricConfig{},
		ActiveNamespaces:         map[string]NamespaceMetadata{},
		SingleNodeMode:           singleNodeMode,
	}
	return nsm
}

func (nsm *NamespaceManager) SetNamespaceFromData(data []byte) {
	var doc map[string]interface{}
	json.Unmarshal(data, &doc)
	// Extract namespace
	ns := doc["namespace"].(string)

	// Extract event configs
	ecs := extractEventConfigs(doc)

	// Extract metric configs
	mcs := extractMetricConfigs(doc, nsm.storage)

	nsm.SetNamespaceFromConfig(ns, ecs, mcs)
}

func (nsm *NamespaceManager) SetNamespaceFromConfig(ns string, ecs map[int]*eventConfig, mcs map[int]*metricConfig) {
	// Create a map from event id to metric configs
	nsMap := make(map[int][]*metricConfig)

	for _, mc := range mcs {
		for _, eventID := range mc.EventIds {
			// Initialize the slice if it doesn't exist
			_, metricExists := nsMap[eventID]
			if !metricExists {
				nsMap[eventID] = []*metricConfig{}
			}
			nsMap[eventID] = append(nsMap[eventID], mc)
		}
	}
	nsm.EventToMetricMap[ns] = nsMap

	// Set map from event id to event config
	nsm.EventConfigsByNamespace[ns] = ecs

	// Set map from metric ID to metric
	nsm.MetricConfigsByNamespace[ns] = mcs

	if nsm.SingleNodeMode {
		nsm.ActivateNamespace(ns)
	}
}

func (nsm *NamespaceManager) ActivateNamespace(ns string) {
	nsm.NsDataLck.Lock()
	nsLogger.Info().Msgf("Activating namespace %s", ns)
	if _, exists := nsm.ActiveNamespaces[ns]; exists {
		nsm.NsDataLck.Unlock()
		return
	}

	nsm.ActiveNamespaces[ns] = NamespaceMetadata{
		KeySizeMap: map[int]int{},
	}

	for _, mc := range nsm.MetricConfigsByNamespace[ns] {
		// TODO init with old values
		nsm.ActiveNamespaces[ns].KeySizeMap[mc.ID] = 0

	}
	nsm.NsDataLck.Unlock()
}

func (nsm *NamespaceManager) DeactivateNamespace(ns string) {
	nsm.NsDataLck.Lock()
	delete(nsm.ActiveNamespaces, ns)
	nsm.NsDataLck.Unlock()
}

func extractEventConfigs(doc map[string]interface{}) map[int]*eventConfig {
	reConfigs := doc["events"].([]interface{})
	eConfigs := map[int]*eventConfig{}
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
		eConfigs[ec.ID] = &ec
	}
	return eConfigs
}

func extractMetricConfigs(doc map[string]interface{}, storage AbstractStorage) map[int]*metricConfig {
	rmConfigs := doc["metrics"].([]interface{})
	mConfigs := map[int]*metricConfig{}
	for _, rmConfig := range rmConfigs {
		rmConfig := rmConfig.(map[string]interface{})
		// Get the initializers for the metric config
		name := rmConfig["name"].(string)
		id := int(rmConfig["id"].(float64))
		keyField := rmConfig["key_field"].(string)
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
			Filter:     extractMetricFilters(rmConfig["filter"].([]interface{})),
			Storage:    storage,
		}
		mConfigs[mc.ID] = &mc
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
