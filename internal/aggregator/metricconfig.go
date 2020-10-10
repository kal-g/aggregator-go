package aggregator

import (
	"errors"
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

// MetricCountResult is the result of a count operation
// TODO move
type MetricCountResult struct {
	Err   error
	Count int
}

func (mc metricConfig) handleEvent(event event) (metricHandleResult, bool) {
	isNew := false
	// Can assume that event is of the right type
	// Check that the event passes the filter
	if !mc.Filter.IsValid(event) {
		return failedFilter, false
	}

	// Get metric from storage, or initialize if it doesn't exist
	mc.Storage.Lock(mc.Namespace)
	storageKey := getMetricStorageKey(event.GetDataField(mc.KeyField).(int), mc.ID, mc.Namespace)
	r := mc.Storage.Get(storageKey)

	if errors.Is(r.Err, &StorageKeyNotFoundError{}) {
		isNew = true
	}

	// Determine how much to increment metric by
	incrementBy := int(1)
	if mc.CountField != "" {
		incrementBy = event.GetDataField(mc.CountField).(int)
	}

	// Put back in storage
	mc.Storage.IncrBy(storageKey, incrementBy)
	mc.Storage.Unlock(mc.Namespace)
	return noError, isNew

}

func (mc metricConfig) getCount(metricKey int) MetricCountResult {
	storageKey := getMetricStorageKey(metricKey, mc.ID, mc.Namespace)
	r := mc.Storage.Get(storageKey)
	if r.Err != nil {
		return MetricCountResult{
			Err:   r.Err,
			Count: 0,
		}
	}
	return MetricCountResult{
		Err:   nil,
		Count: r.Value,
	}
}

func getMetricStorageKey(key int, metricID int, namespace string) string {
	// Key of a metric is the id + the type + the key field
	mk := namespace + ":" + strconv.Itoa(metricID) + ":" + strconv.Itoa(key)
	return mk
}
