package aggregator

// NullFilter passes all events
type NullFilter struct{}

// IsValid passes all events
func (f NullFilter) IsValid(e event) bool {
	return true
}
