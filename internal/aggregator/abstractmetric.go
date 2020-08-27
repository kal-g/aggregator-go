package aggregator

type abstractMetric interface {
	Increment(int) metricHandleResult
	GetValue() int
}
