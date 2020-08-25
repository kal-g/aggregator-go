package aggregator

type CountMetric struct {
	count int
}

func (m *CountMetric) Increment(val int) MetricHandleResult {
	m.count += val
	return NoError
}

func (m CountMetric) GetValue() int {
	return m.count
}
