package aggregator

// Engine is the core calculcation engine for counting
type Engine struct {
	Nsm *NamespaceManager
}

// NewEngine creates a new engine
func NewEngine(nsm *NamespaceManager) Engine {
	e := Engine{Nsm: nsm}
	return e
}

// HandleRawEvent is the handler for any event. It checks whether the event is defined in config
// and updates the relevant metrics
func (e Engine) HandleRawEvent(rawEvent map[string]interface{}, namespace string) EngineHandleResult {
	// Event must have an id to identify what event it is
	id, idExists := rawEvent["id"]
	if !idExists {
		return InvalidEventID
	}
	// Id must be an int
	idTyped, isInt := id.(int)
	if !isInt {
		return InvalidEventID
	}
	// Get the config for the event
	// TODO Add RW lock for events
	eventConfig, configExists := e.Nsm.EventMap[idTyped]
	if !configExists {
		return EventConfigNotFound
	}
	// Validate against the config
	event := eventConfig.validate(rawEvent)
	if event == nil {
		return EventValidationFailed
	}
	res := e.handleEvent(*event, namespace)
	return res
}

func (e Engine) handleEvent(event event, namespace string) EngineHandleResult {
	// TODO Figure out more elegant solution for global + namespace
	e.Nsm.namespaceRLock("")
	if namespace != "" {
		e.Nsm.namespaceRLock(namespace)
	}
	// Get the metric configs for this event
	metricConfigs := e.getMetricConfigs(event, namespace)
	if len(metricConfigs) == 0 {
		return NoMetricsFound
	}

	// Handle this event for each of these metric configs
	for _, metricConfig := range metricConfigs {
		_, isNew := metricConfig.handleEvent(event)
		if isNew {
			e.Nsm.MetaMtx.Lock()
			e.Nsm.NsMetaMap[metricConfig.Namespace].KeySizeMap[metricConfig.ID]++
			e.Nsm.MetaMtx.Unlock()
		}
	}
	e.Nsm.namespaceRUnlock("")
	if namespace != "" {
		e.Nsm.namespaceRUnlock(namespace)
	}
	return Success
}

func (e Engine) getMetricConfigs(event event, namespace string) []*metricConfig {
	configs := []*metricConfig{}

	// First get all configs in global namespace
	globalNamespace, globalNamespaceExists := e.Nsm.EventToMetricMap[""]
	if globalNamespaceExists {
		globalConfigs, globalConfigsExist := globalNamespace[event.ID]
		if globalConfigsExist {
			configs = append(configs, globalConfigs...)
		}
	}

	// Then get all configs in the specified namespace
	if namespace != "" {
		specificNamespace, namespaceExists := e.Nsm.EventToMetricMap[namespace]
		if namespaceExists {
			namespaceConfigs, namespaceConfigsExist := specificNamespace[event.ID]
			if namespaceConfigsExist {
				configs = append(configs, namespaceConfigs...)
			}
		}
	}
	return configs
}

// GetMetricCount gets the value for a given metric
func (e Engine) GetMetricCount(namespaceName string, metricKey int, metricID int) MetricCountResult {

	namespace, namespaceExists := e.Nsm.MetricMap[namespaceName]
	if !namespaceExists {
		return MetricCountResult{
			ErrCode: 1,
			Count:   0,
		}
	}
	mc, mcExists := namespace[metricID]
	if !mcExists {
		return MetricCountResult{
			ErrCode: 2,
			Count:   0,
		}
	}
	return mc.getCount(metricKey)
}
