package configmutator

func (cm ConfigMutator) GetConfig() map[string]interface{} {
	events := []EventConfig{}
	metrics := []MetricConfig{}
	for _, e := range cm.c.Events {
		events = append(events, e)
	}
	for _, m := range cm.c.Metrics {
		metrics = append(metrics, m)
	}
	return map[string]interface{}{
		"namespace":  cm.c.Namespace,
		"metrics":    metrics,
		"events":     events,
		"extra_info": cm.c.ExtraInfo,
	}
}
