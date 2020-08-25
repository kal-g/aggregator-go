package aggregator

type AbstractMetric interface {
	Increment(int) MetricHandleResult
	GetValue() int
}
