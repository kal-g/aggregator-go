package aggregator

import (
	"strconv"
)

type metricConfig struct {
	ID         int
	Name       string
	EventIds   []int
	KeyField   string
	CountField string
	MetricType metricType
	Namespace  string
	Filter     abstractFilter
	Storage    AbstractStorage
}

type MetricCountResult struct {
	ErrCode int
	Count   int
}

func (mc metricConfig) handleEvent(event event) metricHandleResult {

	// Can assume that event is of the right type
	// Check that the event passes the filter
	if !mc.Filter.IsValid(event) {
		return failedFilter
	}

	// Get metric from storage, or initialize if it doesn't exist
	mc.Storage.Lock(mc.Namespace)

	storageKey := getMetricStorageKey(event.GetDataField(mc.KeyField).(int), mc.ID, mc.Namespace)
	r := mc.Storage.Get(storageKey)

	initialValue := int(0)
	if r.ErrCode == 0 {
		initialValue = r.Value
	}

	metric := mc.initMetricByType(initialValue)

	// Determine how much to increment metric by
	incrementBy := int(1)
	if mc.CountField != "" {
		incrementBy = event.GetDataField(mc.CountField).(int)
	}

	// Increment the metric
	metric.Increment(incrementBy)

	// Put back in storage
	mc.Storage.Put(storageKey, metric.GetValue())
	mc.Storage.Unlock(mc.Namespace)
	return noError

}

func (mc metricConfig) getCount(metricKey int) MetricCountResult {
	storageKey := getMetricStorageKey(metricKey, mc.ID, mc.Namespace)
	r := mc.Storage.Get(storageKey)
	if r.ErrCode != 0 {
		return MetricCountResult{
			ErrCode: r.ErrCode,
			Count:   0,
		}
	}
	return MetricCountResult{
		ErrCode: 0,
		Count:   r.Value,
	}
}

func getMetricStorageKey(key int, metricID int, namespace string) string {
	// Key of a metric is the id + the type + the key field
	mk := namespace + ":" + strconv.Itoa(metricID) + ":" + strconv.Itoa(key)
	return mk
}

func (mc metricConfig) initMetricByType(val int) abstractMetric {
	return &countMetric{
		count: val,
	}
}
