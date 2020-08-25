package aggregator

type MetricType int32
type FieldType int32
type EngineHandleResult int32
type MetricHandleResult int32

const (
	CountMetricType MetricType = 0
)

const (
	StringField FieldType = 0
	IntField    FieldType = 1
)

const (
	Success               EngineHandleResult = 0
	NoMetricsFound        EngineHandleResult = 1
	EventValidationFailed EngineHandleResult = 2
	EventConfigNotFound   EngineHandleResult = 3
	InvalidEventId        EngineHandleResult = 4
)

const (
	NoError      MetricHandleResult = 0
	FailedFilter MetricHandleResult = 1
)

type Value interface{}
type RawEvent map[string]Value
