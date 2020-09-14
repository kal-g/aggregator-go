package aggregator

type countMetric struct {
	count int
}

func (m countMetric) GetValue() int {
	return m.count
}
