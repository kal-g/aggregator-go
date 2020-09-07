package aggregator

type metricType int32
type fieldType int32
type EngineHandleResult int32
type metricHandleResult int32

const (
	countMetricType metricType = 0
)

const (
	stringField fieldType = 0
	intField    fieldType = 1
)

const (
	Success               EngineHandleResult = 0
	NoMetricsFound        EngineHandleResult = 1
	EventValidationFailed EngineHandleResult = 2
	EventConfigNotFound   EngineHandleResult = 3
	InvalidEventID        EngineHandleResult = 4
	DeferredSuccess       EngineHandleResult = 5
)

const (
	noError      metricHandleResult = 0
	failedFilter metricHandleResult = 1
)

type value interface{}
type rawEvent map[string]value
