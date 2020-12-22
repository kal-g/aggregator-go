package aggregator

import "fmt"

// Storage errors
type StorageKeyNotFoundError struct{}

func (e *StorageKeyNotFoundError) Error() string {
	return "Storage - Key not found"
}

type StorageInvalidDataError struct{}

func (e *StorageInvalidDataError) Error() string {
	return "Storage - Invalid Data"
}

// Engine Errors
type InvalidEventIDError struct{}

func (e *InvalidEventIDError) Error() string {
	return "Invalid event ID"
}

type EventConfigNotFoundError struct{}

func (e *EventConfigNotFoundError) Error() string {
	return "Event config not found"
}

type EventValidationFailedError struct {
	msg string
}

func (e *EventValidationFailedError) Error() string {
	return fmt.Sprintf("Event validation failed: %+v", e.msg)
}

type NoMetricsFoundError struct{}

func (e *NoMetricsFoundError) Error() string {
	return "No metrics found"
}

// Other errors

type NamespaceExistsError struct{}

func (e *NamespaceExistsError) Error() string {
	return "Namespace exists"
}

// TODO Make this print relative file and line numbers when debug mode activated
type NamespaceNotFoundError struct{}

func (e *NamespaceNotFoundError) Error() string {
	return "Namespace not found"
}

type MetricConfigNotFoundError struct{}

func (e *MetricConfigNotFoundError) Error() string {
	return "Metric config not found"
}

type MetricKeyNotFoundError struct{}

func (e *MetricKeyNotFoundError) Error() string {
	return "Metric key not found"
}

type MetricIDNotFoundError struct{}

func (e *MetricIDNotFoundError) Error() string {
	return "Metric ID not found"
}

type MetricKeyInvalidType struct{}

func (e *MetricKeyInvalidType) Error() string {
	return "Metric key invalid type"
}

type MetricIDInvalidType struct{}

func (e *MetricIDInvalidType) Error() string {
	return "Metric ID invalid type"
}
