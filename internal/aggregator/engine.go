package aggregator

import (
	"os"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger = zerolog.New(os.Stderr).With().
	Str("source", "ENG").
	Timestamp().
	Logger()

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
func (e Engine) HandleRawEvent(rawEvent map[string]interface{}, namespace string) error {
	// Event must have an id to identify what event it is
	id, idExists := rawEvent["id"]
	if !idExists {
		return &InvalidEventIDError{}
	}
	// Id must be an int
	idTyped, isInt := id.(int)
	if !isInt {
		return &InvalidEventIDError{}
	}
	// Get the config for the event
	// TODO Add RW lock for events
	eventConfig, configExists := e.Nsm.EventConfigsByNamespace[namespace][idTyped]
	if !configExists {
		return &EventConfigNotFoundError{}
	}
	// Validate against the config
	event := eventConfig.validate(rawEvent)
	if event == nil {
		return &EventValidationFailedError{}
	}
	// Check namespace
	e.Nsm.NsDataLck.RLock()
	if _, nsExists := e.Nsm.ActiveNamespaces[namespace]; !nsExists {
		return &NamespaceNotFoundError{}
	}
	e.Nsm.NsDataLck.RUnlock()
	res := e.handleEvent(*event, namespace)
	return res
}

func (e Engine) handleEvent(event event, namespace string) error {
	e.Nsm.namespaceRLock(namespace)
	// Get the metric configs for this event
	metricConfigs := e.getMetricConfigs(event, namespace)
	if len(metricConfigs) == 0 {
		return &NoMetricsFoundError{}
	}

	// Handle this event for each of these metric configs
	for _, metricConfig := range metricConfigs {
		_, isNew := metricConfig.handleEvent(event, namespace)
		if isNew {
			e.Nsm.NsDataLck.RLock()
			e.Nsm.ActiveNamespaces[namespace].KeySizeMap[metricConfig.ID]++
			e.Nsm.NsDataLck.RUnlock()
		}
	}
	e.Nsm.namespaceRUnlock(namespace)
	return nil
}

func (e Engine) getMetricConfigs(event event, namespace string) []*metricConfig {
	configs := []*metricConfig{}
	e.Nsm.NsDataLck.RLock()
	// Get all configs in the specified namespace
	// Check if namespace active on this node
	if _, exists := e.Nsm.ActiveNamespaces[namespace]; exists {
		specificNamespace, namespaceExists := e.Nsm.EventToMetricMap[namespace]
		if namespaceExists {
			namespaceConfigs, namespaceConfigsExist := specificNamespace[event.ID]
			if namespaceConfigsExist {
				configs = append(configs, namespaceConfigs...)
			}
		}
	}
	e.Nsm.NsDataLck.RUnlock()
	return configs
}

// GetMetricCount gets the value for a given metric
func (e Engine) GetMetricCount(ns string, metricKey int, metricID int) MetricCountResult {

	mcs, namespaceExists := e.Nsm.MetricConfigsByNamespace[ns]
	if !namespaceExists {
		return MetricCountResult{
			Err:   &NamespaceNotFoundError{},
			Count: 0,
		}
	}
	mc, mcExists := mcs[metricID]
	if !mcExists {
		return MetricCountResult{
			Err:   &MetricConfigNotFoundError{},
			Count: 0,
		}
	}
	return mc.getCount(metricKey, ns)
}
