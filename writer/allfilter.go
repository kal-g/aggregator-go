package aggregator

type AllFilter struct {
  Filters []AbstractFilter
}

func (f AllFilter) IsValid(e Event) bool {
  for _, filter := range f.Filters {
    if (!filter.IsValid(e)) {
      return false
    }
  }
  return true
}
