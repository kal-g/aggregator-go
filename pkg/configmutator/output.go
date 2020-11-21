package configmutator

func (cm ConfigMutator) GetConfig() map[string]interface{} {
	events := []EventConfig{}
	metrics := []MetricConfig{}
	for _, e := range cm.C.Events {
		events = append(events, e)
	}
	for _, m := range cm.C.Metrics {
		metrics = append(metrics, m)
	}
	return map[string]interface{}{
		"namespace":  cm.C.Namespace,
		"metrics":    metrics,
		"events":     events,
		"extra_info": cm.C.ExtraInfo,
	}
}
