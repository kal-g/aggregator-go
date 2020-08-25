package aggregator

type EventConfig struct {
	Name   string
	Id     int
	Fields map[string]FieldType
}

// TODO Add logging on nil returns here
func (ec EventConfig) Validate(re map[string]interface{}) *Event {
	// We have to have an id, it has to be an int, and it has to match
	idRaw, hasId := re["id"]
	idTyped := 0
	isInt := false
	if !hasId {
		return nil
	} else {
		idTyped, isInt = idRaw.(int)
		if !isInt {
			return nil
		}
		if idTyped != ec.Id {
			return nil
		}
	}
	// Make sure number of fields matches
	// The raw event includes id, so it should have an extra field
	if len(re) != len(ec.Fields)+1 {
		return nil
	}
	// Iterate over each field in the config, and make sure type matches config
	for fieldName, fieldType := range ec.Fields {
		// First check the field exists
		fieldTypeRaw, fieldExists := re[fieldName]
		if !fieldExists {
			return nil
		}
		// If it exists, check the type matches
		switch fieldType {
		case StringField:
			_, isString := fieldTypeRaw.(string)
			if !isString {
				return nil
			}
		case IntField:
			_, isInt := fieldTypeRaw.(int)
			if !isInt {
				return nil
			}
		}
	}
	delete(re, "id")
	return &Event{
		Id:   idTyped,
		Data: re,
	}
}
