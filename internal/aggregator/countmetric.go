package aggregator

type countMetric struct {
	count int
}

func (m *countMetric) Increment(val int) metricHandleResult {
	m.count += val
	return noError
}

func (m countMetric) GetValue() int {
	return m.count
}
