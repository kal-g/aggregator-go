package configmutator

type EventNotFoundError struct{}

func (e *EventNotFoundError) Error() string {
	return "Event not found"
}

type FieldTypeConflict struct{}

func (e *FieldTypeConflict) Error() string {
	return "Field type conflict"
}

type FieldAlreadyExists struct{}

func (e *FieldAlreadyExists) Error() string {
	return "Field already exists"
}

type InvalidKeyField struct{}

func (e *InvalidKeyField) Error() string {
	return "Invalid key field"
}

type InvalidCountField struct{}

func (e *InvalidCountField) Error() string {
	return "Invalid count field"
}
