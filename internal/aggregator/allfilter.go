package aggregator

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
