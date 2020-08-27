package aggregator

type metricType int32
type fieldType int32
type engineHandleResult int32
type metricHandleResult int32

const (
	countMetricType metricType = 0
)

const (
	stringField fieldType = 0
	intField    fieldType = 1
)

const (
	success               engineHandleResult = 0
	noMetricsFound        engineHandleResult = 1
	eventValidationFailed engineHandleResult = 2
	eventConfigNotFound   engineHandleResult = 3
	invalidEventID        engineHandleResult = 4
	deferredSuccess       engineHandleResult = 5
)

const (
	noError      metricHandleResult = 0
	failedFilter metricHandleResult = 1
)

type value interface{}
type rawEvent map[string]value
