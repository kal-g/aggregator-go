package configmutator

import "encoding/json"

func (cm *ConfigMutator) AddNewMetric(name string, kf string, cf string) error {
	// Check for valid key field
	if _, exists := cm.KeyFields[kf]; !exists {
		return &InvalidKeyField{}
	}
	if _, exists := cm.CountFields[cf]; cf != "" && !exists {
		return &InvalidCountField{}
	}
	// Check for valid count field
	filter := []interface{}{"null"}
	filterString, _ := json.Marshal(filter)
	cm.C.Metrics[cm.nextMetricID] = MetricConfig{
		ID:           cm.nextMetricID,
		Name:         name,
		EventIDs:     []int{},
		KeyField:     kf,
		CountField:   cf,
		Type:         "count",
		Filter:       filter,
		FilterString: string(filterString),
	}
	cm.nextMetricID++
	cm.Update()
	return nil
}

func (cm *ConfigMutator) AddEventID(metricID int, eventID int) error {
	// Check if metric exists
	if _, exists := cm.C.Metrics[metricID]; !exists {
		return &MetricNotFoundError{}
	}
	// Check if event ID exists
	for _, id := range cm.C.Metrics[metricID].EventIDs {
		if eventID == id {
			return &EventIDExists{}
		}
	}

	// Add event id
	m := cm.C.Metrics[metricID]
	m.EventIDs = append(m.EventIDs, eventID)
	cm.C.Metrics[metricID] = m
	cm.Update()
	return nil
}

func (cm *ConfigMutator) SetFilterForMetric(id int, filter []interface{}) {
	m := cm.C.Metrics[id]
	m.Filter = filter
	cm.C.Metrics[id] = m
}

func (cm *ConfigMutator) GetNewEventIDsForMetric(id int) []int {
	keyField := cm.C.Metrics[id].KeyField
	countField := cm.C.Metrics[id].CountField
	eventIDs := cm.C.Metrics[id].EventIDs
	eventIDsMap := map[int]bool{}
	for _, id := range eventIDs {
		eventIDsMap[id] = true
	}
	// Need to get a list of ids that contain both the keyField and the countField,
	// and that are not already in the event ids for the metric
	keyFieldIDs := cm.fieldToEventIDs[keyField]
	countFieldIDs := cm.fieldToEventIDs[countField]

	newIDs := []int{}
	for id := range keyFieldIDs {
		_, cfExists := countFieldIDs[id]
		_, idExists := eventIDsMap[id]
		if (countField == "" || cfExists) && !idExists {
			newIDs = append(newIDs, id)
		}
	}

	// TODO filter

	return newIDs
}
