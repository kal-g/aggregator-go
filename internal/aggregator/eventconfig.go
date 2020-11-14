package aggregator

type eventConfig struct {
	Name   string
	ID     int
	Fields map[string]fieldType
}

// TODO Add logging on nil returns here
// Validate makes sure that this is a valid event, based on names and types
// After conversion to event, typing is guaranteed
func (ec eventConfig) validate(re map[string]interface{}) *event {
	// We have to have an id, it has to be an int, and it has to match
	idRaw, hasID := re["id"]
	idTyped := 0
	isInt := false
	if !hasID {
		return nil
	}
	idTyped, isInt = idRaw.(int)
	if !isInt {
		return nil
	}
	if idTyped != ec.ID {
		return nil
	}
	// Make sure number of fields matches
	// The raw event includes id, so it should have an extra field
	if len(re) != len(ec.Fields)+1 {
		return nil
	}
	logger.Info().Msgf("Fields: %+v", ec.Fields)
	// Iterate over each field in the config, and make sure type matches config
	for fieldName, fieldType := range ec.Fields {
		// First check the field exists
		fieldTypeRaw, fieldExists := re[fieldName]
		if !fieldExists {
			return nil
		}
		// If it exists, check the type matches
		switch fieldType {
		case stringField:
			_, isString := fieldTypeRaw.(string)
			if !isString {
				return nil
			}
		case intField:
			_, isInt := fieldTypeRaw.(int)
			if !isInt {
				return nil
			}
		default:
			panic("Invalid field type")
		}

	}
	delete(re, "id")
	return &event{
		ID:   idTyped,
		Data: re,
	}
}
