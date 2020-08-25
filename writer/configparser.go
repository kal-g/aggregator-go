package aggregator

import "encoding/json"

type ConfigParser struct {
	EventConfigs  []EventConfig
	MetricConfigs []MetricConfig
	storage       AbstractStorage
}

func NewConfigParserFromRaw(input_string []byte, storage AbstractStorage) ConfigParser {
	var doc map[string]interface{}
	json.Unmarshal(input_string, &doc)

	return ConfigParser{extractEventConfigs(doc), extractMetricConfigs(doc, storage), storage}
}

func NewConfigParserFromConfigs(event_configs []EventConfig, metric_configs []MetricConfig, storage AbstractStorage) ConfigParser {
	return ConfigParser{
		EventConfigs:  event_configs,
		MetricConfigs: metric_configs,
		storage:       storage,
	}
}

func (cp ConfigParser) GetEventConfigs() []EventConfig {
	return cp.EventConfigs
}

func (cp ConfigParser) GetMetricConfigs() []MetricConfig {
	return cp.MetricConfigs
}

func extractEventConfigs(doc map[string]interface{}) []EventConfig {
	raw_event_configs := doc["events"].([]interface{})
	event_configs := []EventConfig{}
	for _, raw_event_config := range raw_event_configs {
		raw_event_config := raw_event_config.(map[string]interface{})
		// Get the initializers for the event config
		name := raw_event_config["name"].(string)
		id := int(raw_event_config["id"].(float64))
		raw_fields := raw_event_config["fields"].(map[string]interface{})
		fields := map[string]FieldType{}
		for field_name, field_type := range raw_fields {
			fields[field_name] = FieldType(field_type.(float64))
		}
		// Create event config
		event_config := EventConfig{
			Name:   name,
			Id:     id,
			Fields: fields,
		}
		event_configs = append(event_configs, event_config)
	}
	return event_configs
}

func extractMetricConfigs(doc map[string]interface{}, storage AbstractStorage) []MetricConfig {
	raw_metric_configs := doc["metrics"].([]interface{})
	metric_configs := []MetricConfig{}
	for _, raw_metric_config := range raw_metric_configs {
		raw_metric_config := raw_metric_config.(map[string]interface{})
		// Get the initializers for the metric config
		name := raw_metric_config["name"].(string)
		id := int(raw_metric_config["id"].(float64))
		key_field := raw_metric_config["key_field"].(string)
		//metric_type := raw_metric_config["type"].(string)
		count_field := raw_metric_config["count_field"].(string)
		raw_event_ids := raw_metric_config["event_ids"].([]interface{})
		event_ids := []int{}
		for _, event_id := range raw_event_ids {
			event_ids = append(event_ids, int(event_id.(float64)))
		}
		// Create metric config
		metric_config := MetricConfig{
			Id:         id,
			Name:       name,
			EventIds:   event_ids,
			KeyField:   key_field,
			CountField: count_field,
			MetricType: CountMetricType,
			Filter:     extractMetricFilters(raw_metric_config["filter"].([]interface{})),
			Storage:    storage,
		}
		metric_configs = append(metric_configs, metric_config)
	}
	return metric_configs
}

func extractMetricFilters(filt []interface{}) AbstractFilter {
	// Name is always the first element
	var f AbstractFilter
	filter_name := filt[0].(string)
	if filter_name == "null" {
		f = NullFilter{}
	} else if filter_name == "gt" {
		f = GreaterThanFilter{filt[1].(string), int(filt[2].(float64))}
	} else if filter_name == "all" {
		filters := []AbstractFilter{}
		for i := 1; i < len(filt); i++ {
			filters = append(filters, extractMetricFilters(filt[i].([]interface{})))
		}
		f = AllFilter{filters}
	}
	return f
}
