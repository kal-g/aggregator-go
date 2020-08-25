package aggregator

import "fmt"

type Engine struct {
	EventMap  map[int]EventConfig
	MetricMap map[int][]MetricConfig
	Parser    *ConfigParser
}

func NewEngine(parser *ConfigParser) Engine {
	event_map := make(map[int]EventConfig)
	// Create a map from event id to event config
	for _, event_config := range parser.GetEventConfigs() {
		event_map[event_config.Id] = event_config
	}

	metric_map := make(map[int][]MetricConfig)
	// Create a map from event id to metric configs
	for _, metric_config := range parser.GetMetricConfigs() {
		for _, event_id := range metric_config.EventIds {
			// Initialize the slice if it doesn't exist
			_, exists := metric_map[event_id]
			if !exists {
				metric_map[event_id] = []MetricConfig{}
			}
			metric_map[event_id] = append(metric_map[event_id], metric_config)
		}
	}

	return Engine{
		EventMap:  event_map,
		MetricMap: metric_map,
		Parser:    parser,
	}
}

func (e Engine) HandleRawEvent(raw_event map[string]interface{}) EngineHandleResult {
	// Event must have an id to identify what event it is
	id, id_exists := raw_event["id"]
	if !id_exists {
		fmt.Printf("Here 1\n")
		return InvalidEventId
	}
	// Id must be an int
	id_typed, is_int := id.(int)
	fmt.Printf("Thing %d\n", id)
	if !is_int {
		fmt.Printf("Here 2\n")
		return InvalidEventId
	}
	// Get the config for the event
	event_config, config_exists := e.EventMap[id_typed]
	if !config_exists {
		return EventConfigNotFound
	}
	// Validate against the config
	event := event_config.Validate(raw_event)
	if event == nil {
		return EventValidationFailed
	}
	return e.handleEvent(*event)
}

func (e Engine) handleEvent(event Event) EngineHandleResult {
	// Get the metric configs for this event
	metric_configs, config_exists := e.MetricMap[event.Id]
	if !config_exists {
		return NoMetricsFound
	}

	// Handle this event for each of these metric configs
	for _, metric_config := range metric_configs {
		metric_config.HandleEvent(event)
	}
	return Success
}
