package aggregator

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
