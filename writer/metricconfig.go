package aggregator

import (
	"strconv"
)

type MetricConfig struct {
	Id         int
	Name       string
	EventIds   []int
	KeyField   string
	CountField string
	MetricType MetricType
	Filter     AbstractFilter
	Storage    AbstractStorage
}

func (mc MetricConfig) HandleEvent(event Event) MetricHandleResult {

	// Can assume that event is of the right type
	// Check that the event passes the filter
	if !mc.Filter.IsValid(event) {
		return FailedFilter
	}

	// Get metric from storage, or initialize if it doesn't exist
	mc.Storage.Lock()

	storage_key := mc.getMetricKey(event)
	r := mc.Storage.Get(storage_key)

	initial_value := int(0)
	if r.ErrCode == 0 {
		initial_value = r.Value
	}

	metric := mc.initMetricByType(initial_value)

	// Determine how much to increment metric by
	increment_by := int(1)
	if mc.CountField != "" {
		increment_by = event.GetDataField(mc.CountField).(int)
	}

	// Increment the metric
	metric.Increment(increment_by)

	// Put back in storage
	mc.Storage.Put(storage_key, metric.GetValue())
	mc.Storage.Unlock()
	return NoError

}

func (mc MetricConfig) getMetricKey(event Event) string {
	// Key of a metric is the id + the type + the key field
	return strconv.Itoa(mc.Id) + ":" + strconv.Itoa(event.GetDataField(mc.KeyField).(int))
}

func (mc MetricConfig) initMetricByType(val int) AbstractMetric {
	return &CountMetric{
		count: val,
	}
}
