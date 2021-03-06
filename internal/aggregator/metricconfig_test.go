package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
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
	res1, _ := config.handleEvent(validEvent, "test")
	res2, _ := config.handleEvent(invalidEvent, "test")
	ct.AssertEqual(t, res1, noError)
	ct.AssertEqual(t, res2, failedFilter)

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

	config.handleEvent(event, "test")

	// Metric 1, for key 1234
	sr := storage.Get("test:1:1234")

	ct.AssertEqual(t, sr.Err, nil)
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

	config.handleEvent(event, "test")
	sr1 := storage.Get("test:1:1234")
	ct.AssertEqual(t, sr1.Err, nil)
	ct.AssertEqual(t, sr1.Value, 1)

	config.handleEvent(event, "test")
	sr2 := storage.Get("test:1:1234")
	ct.AssertEqual(t, sr2.Err, nil)
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

	config.handleEvent(e1, "test")
	sr1 := storage.Get("test:1:1234")
	ct.AssertEqual(t, sr1.Err, nil)
	ct.AssertEqual(t, sr1.Value, 5)

	config.handleEvent(e2, "test")
	sr2 := storage.Get("test:1:1234")
	ct.AssertEqual(t, sr2.Err, nil)
	ct.AssertEqual(t, sr2.Value, 3)
}
