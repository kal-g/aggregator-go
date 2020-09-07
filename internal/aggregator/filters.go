package aggregator

type abstractFilter interface {
	IsValid(event) bool
}

type allFilter struct {
	Filters []abstractFilter
}

func (f allFilter) IsValid(e event) bool {
	for _, filter := range f.Filters {
		if !filter.IsValid(e) {
			return false
		}
	}
	return true
}

type greaterThanFilter struct {
	Key   string
	Value int
}

func (f greaterThanFilter) IsValid(e event) bool {
	eVal, ok := e.GetDataField(f.Key).(int)
	if !ok {
		panic("Invalid data field")
	}
	if eVal > f.Value {
		return true
	}
	return false
}

// NullFilter passes all events
type NullFilter struct{}

// IsValid passes all events
func (f NullFilter) IsValid(e event) bool {
	return true
}
