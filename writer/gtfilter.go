package aggregator

type GreaterThanFilter struct {
	Key   string
	Value int
}

func (f GreaterThanFilter) IsValid(e Event) bool {
	event_value, ok := e.GetDataField(f.Key).(int)
	if !ok {
		panic("Invalid data field")
	}
	if event_value > f.Value {
		return true
	} else {
		return false
	}
}
