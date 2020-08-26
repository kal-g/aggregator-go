package aggregator

type event struct {
	ID   int
	Data map[string]interface{}
}

func (e event) GetDataField(field string) value {
	return e.Data[field]
}
