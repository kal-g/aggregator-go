package configmutator

func (cm *ConfigMutator) AddNewMetric(name string, kf string, cf string) error {
	// Check for valid key field
	if _, exists := cm.KeyFields[kf]; !exists {
		return &InvalidKeyField{}
	}
	if _, exists := cm.CountFields[cf]; cf != "" && !exists {
		return &InvalidCountField{}
	}
	// Check for valid count field
	cm.C.Metrics[cm.nextMetricID] = MetricConfig{
		ID:         cm.nextMetricID,
		Name:       name,
		EventIDs:   []int{},
		KeyField:   kf,
		CountField: cf,
		Type:       "count",
		Filter:     []interface{}{"null"},
	}
	cm.nextMetricID++
	cm.Update()
	return nil
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
