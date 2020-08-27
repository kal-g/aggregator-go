package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/common_test"
)

func TestFilterValidation(t *testing.T) {
	eIDs := []int{1}
	filter := greaterThanFilter{
		Key:   "filterKey",
		Value: 0,
	}
	storage := newNaiveStorage()
	config := metricConfig{
		ID:         1,
		Name:       "testConfig",
		EventIds:   eIDs,
		KeyField:   "key",
		CountField: "",
		MetricType: countMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create events
	validEvent := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234, "filterKey": 1},
	}
	invalidEvent := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234, "filterKey": -1},
	}

	ct.AssertEqual(t, config.handleEvent(validEvent), noError)
	ct.AssertEqual(t, config.handleEvent(invalidEvent), failedFilter)

}

func TestMetricStorageInit(t *testing.T) {
	eIDs := []int{1}
	filter := NullFilter{}
	storage := newNaiveStorage()
	config := metricConfig{
		ID:         1,
		Name:       "testConfig",
		EventIds:   eIDs,
		KeyField:   "key",
		CountField: "",
		MetricType: countMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create event
	event := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234},
	}

	config.handleEvent(event)

	// Metric 1, for key 1234
	sr := storage.Get(":1:1234")

	ct.AssertEqual(t, sr.ErrCode, 0)
	ct.AssertEqual(t, sr.Value, 1)
}

func TestMetricStateMaintained(t *testing.T) {
	eIDs := []int{1}
	filter := NullFilter{}
	storage := newNaiveStorage()
	config := metricConfig{
		ID:         1,
		Name:       "testConfig",
		EventIds:   eIDs,
		KeyField:   "key",
		CountField: "",
		MetricType: countMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create event
	event := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234},
	}

	config.handleEvent(event)
	sr1 := storage.Get(":1:1234")
	ct.AssertEqual(t, sr1.ErrCode, 0)
	ct.AssertEqual(t, sr1.Value, 1)

	config.handleEvent(event)
	sr2 := storage.Get(":1:1234")
	ct.AssertEqual(t, sr2.ErrCode, 0)
	ct.AssertEqual(t, sr2.Value, 2)

}

func TestIncDecWithCountKey(t *testing.T) {
	eIDs := []int{1}
	filter := NullFilter{}
	storage := newNaiveStorage()
	config := metricConfig{
		ID:         1,
		Name:       "testConfig",
		EventIds:   eIDs,
		KeyField:   "key",
		CountField: "countKey",
		MetricType: countMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create events
	e1 := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234, "countKey": 5},
	}
	e2 := event{
		ID:   0,
		Data: map[string]interface{}{"key": 1234, "countKey": -2},
	}

	config.handleEvent(e1)
	sr1 := storage.Get(":1:1234")
	ct.AssertEqual(t, sr1.ErrCode, 0)
	ct.AssertEqual(t, sr1.Value, 5)

	config.handleEvent(e2)
	sr2 := storage.Get(":1:1234")
	ct.AssertEqual(t, sr2.ErrCode, 0)
	ct.AssertEqual(t, sr2.Value, 3)
}
