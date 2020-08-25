package aggregator

type Event struct {
	Id   int
	Data map[string]interface{}
}

func (e Event) GetDataField(field string) Value {
	return e.Data[field]
}
