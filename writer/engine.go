package aggregator

type engine struct {
	EventMap  map[int]eventConfig
	MetricMap map[string]map[int][]metricConfig
	Parser    *configParser
}

func newEngine(parser *configParser) engine {
	eventMap := make(map[int]eventConfig)
	// Create a map from event id to event config
	for _, eventConfig := range parser.getEventConfigs() {
		eventMap[eventConfig.ID] = eventConfig
	}

	metricMap := make(map[string]map[int][]metricConfig)
	// Create a map from event id to metric configs
	for _, mc := range parser.getMetricConfigs() {
		// Init the namespace if it doesn't exist
		_, namespaceExists := metricMap[mc.Namespace]
		if !namespaceExists {
			metricMap[mc.Namespace] = make(map[int][]metricConfig)
		}
		for _, eventID := range mc.EventIds {
			// Initialize the slice if it doesn't exist
			_, metricExists := metricMap[mc.Namespace][eventID]
			if !metricExists {
				metricMap[mc.Namespace][eventID] = []metricConfig{}
			}
			metricMap[mc.Namespace][eventID] = append(metricMap[mc.Namespace][eventID], mc)
		}
	}

	return engine{
		EventMap:  eventMap,
		MetricMap: metricMap,
		Parser:    parser,
	}
}

func (e engine) HandleRawEvent(rawEvent map[string]interface{}, namespace string) engineHandleResult {
	// Event must have an id to identify what event it is
	id, idExists := rawEvent["id"]
	if !idExists {
		return invalidEventID
	}
	// Id must be an int
	idTyped, isInt := id.(int)
	if !isInt {
		return invalidEventID
	}
	// Get the config for the event
	eventConfig, configExists := e.EventMap[idTyped]
	if !configExists {
		return eventConfigNotFound
	}
	// Validate against the config
	event := eventConfig.validate(rawEvent)
	if event == nil {
		return eventValidationFailed
	}
	return e.handleEvent(*event, namespace)
}

func (e engine) handleEvent(event event, namespace string) engineHandleResult {
	// Get the metric configs for this event
	metricConfigs := e.getMetricConfigs(event, namespace)
	if len(metricConfigs) == 0 {
		return noMetricsFound
	}

	// Handle this event for each of these metric configs
	for _, metricConfig := range metricConfigs {
		metricConfig.handleEvent(event)
	}
	return success
}

func (e engine) getMetricConfigs(event event, namespace string) []metricConfig {
	configs := []metricConfig{}

	// First get all configs in global namespace
	globalNamespace, globalNamespaceExists := e.MetricMap[""]
	if globalNamespaceExists {
		globalConfigs, globalConfigsExist := globalNamespace[event.ID]
		if globalConfigsExist {
			configs = append(configs, globalConfigs...)
		}
	}

	// Then get all configs in the specified namespace
	if namespace != "" {
		specificNamespace, namespaceExists := e.MetricMap[namespace]
		if namespaceExists {
			namespaceConfigs, namespaceConfigsExist := specificNamespace[event.ID]
			if namespaceConfigsExist {
				configs = append(configs, namespaceConfigs...)
			}
		}
	}
	return configs
}
