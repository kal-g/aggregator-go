package configmutator

type EventNotFoundError struct{}

func (e *EventNotFoundError) Error() string {
	return "Event not found"
}

type MetricNotFoundError struct{}

func (e *MetricNotFoundError) Error() string {
	return "Metric not found"
}

type EventIDExists struct{}

func (e *EventIDExists) Error() string {
	return "Event ID exists"
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
