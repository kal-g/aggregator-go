package aggregator

type NullFilter struct {}

func (f NullFilter) IsValid(e Event) bool {
  return true
}
