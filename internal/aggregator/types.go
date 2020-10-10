package aggregator

type metricType int32
type fieldType int32
type metricHandleResult int32

const (
	countMetricType metricType = 0
)

const (
	stringField fieldType = 0
	intField    fieldType = 1
)

const (
	noError      metricHandleResult = 0
	failedFilter metricHandleResult = 1
)

type value interface{}
type rawEvent map[string]value
