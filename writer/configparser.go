package aggregator

import "encoding/json"

type configParser struct {
	EventConfigs  []eventConfig
	MetricConfigs []metricConfig
	storage       AbstractStorage
}

func newConfigParserFromRaw(input []byte, storage AbstractStorage) configParser {
	var doc map[string]interface{}
	json.Unmarshal(input, &doc)

	return configParser{extractEventConfigs(doc), extractMetricConfigs(doc, storage), storage}
}

func newConfigParserFromConfigs(ecs []eventConfig, mcs []metricConfig, storage AbstractStorage) configParser {
	return configParser{
		EventConfigs:  ecs,
		MetricConfigs: mcs,
		storage:       storage,
	}
}

func (cp configParser) getEventConfigs() []eventConfig {
	return cp.EventConfigs
}

func (cp configParser) getMetricConfigs() []metricConfig {
	return cp.MetricConfigs
}

func extractEventConfigs(doc map[string]interface{}) []eventConfig {
	reConfigs := doc["events"].([]interface{})
	eConfigs := []eventConfig{}
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
		eConfigs = append(eConfigs, ec)
	}
	return eConfigs
}

func extractMetricConfigs(doc map[string]interface{}, storage AbstractStorage) []metricConfig {
	rmConfigs := doc["metrics"].([]interface{})
	mConfigs := []metricConfig{}
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
		mConfigs = append(mConfigs, mc)
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
