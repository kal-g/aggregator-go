package configmutator

func (cm *ConfigMutator) AddNewEvent(name string) {
	cm.C.Events[cm.nextEventID] = EventConfig{
		Name:   name,
		ID:     cm.nextEventID,
		Fields: map[string]int{},
	}
	cm.nextEventID++
	cm.Update()
}

func (cm *ConfigMutator) AddEventField(eID int, fieldName string, fieldType int) error {
	// Check if event exists
	if _, exists := cm.C.Events[eID]; !exists {
		return &EventNotFoundError{}
	}
	// Check for field conflict
	if ft, exists := cm.allFields[fieldName]; exists && (ft != fieldType) {
		return &FieldTypeConflict{}
	}
	// Check if event already contains field
	if _, exists := cm.C.Events[eID].Fields[fieldName]; exists {
		return &FieldAlreadyExists{}
	}
	cm.C.Events[eID].Fields[fieldName] = fieldType
	cm.Update()
	return nil
}
