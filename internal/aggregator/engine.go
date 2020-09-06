package aggregator

type engine struct {
	nsm *namespaceManager
}

func newEngine(nsm *namespaceManager) engine {
	e := engine{nsm: nsm}
	return e
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
	// TODO Add RW lock for events
	eventConfig, configExists := e.nsm.EventMap[idTyped]
	if !configExists {
		return eventConfigNotFound
	}
	// Validate against the config
	event := eventConfig.validate(rawEvent)
	if event == nil {
		return eventValidationFailed
	}
	res := e.handleEvent(*event, namespace)
	return res
}

func (e engine) handleEvent(event event, namespace string) engineHandleResult {
	// TODO Figure out more elegant solution for global + namespace
	e.nsm.namespaceRLock("")
	if namespace != "" {
		e.nsm.namespaceRLock(namespace)
	}
	// Get the metric configs for this event
	metricConfigs := e.getMetricConfigs(event, namespace)
	if len(metricConfigs) == 0 {
		return noMetricsFound
	}

	// Handle this event for each of these metric configs
	for _, metricConfig := range metricConfigs {
		metricConfig.handleEvent(event)
	}
	e.nsm.namespaceRUnlock("")
	if namespace != "" {
		e.nsm.namespaceRUnlock(namespace)
	}
	return success
}

func (e engine) getMetricConfigs(event event, namespace string) []*metricConfig {
	configs := []*metricConfig{}

	// First get all configs in global namespace
	globalNamespace, globalNamespaceExists := e.nsm.EventToMetricMap[""]
	if globalNamespaceExists {
		globalConfigs, globalConfigsExist := globalNamespace[event.ID]
		if globalConfigsExist {
			configs = append(configs, globalConfigs...)
		}
	}

	// Then get all configs in the specified namespace
	if namespace != "" {
		specificNamespace, namespaceExists := e.nsm.EventToMetricMap[namespace]
		if namespaceExists {
			namespaceConfigs, namespaceConfigsExist := specificNamespace[event.ID]
			if namespaceConfigsExist {
				configs = append(configs, namespaceConfigs...)
			}
		}
	}
	return configs
}

func (e engine) getMetricConfig(namespaceName string, metricID int) *metricConfig {
	namespace, namespaceExists := e.nsm.MetricMap[namespaceName]
	if !namespaceExists {
		return nil
	}
	mc, mcExists := namespace[metricID]
	if !mcExists {
		return nil
	}
	return mc
}
