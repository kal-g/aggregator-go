package aggregator_test

import (
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestFilterValidation(t *testing.T) {
	event_ids := []int{1}
	filter := GreaterThanFilter{
		Key:   "filterKey",
		Value: 0,
	}
	storage := NewNaiveStorage()
	config := MetricConfig{
		Id:         1,
		Name:       "testConfig",
		EventIds:   event_ids,
		KeyField:   "key",
		CountField: "",
		MetricType: CountMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create events
	valid_event := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234, "filterKey": 1},
	}
	invalid_event := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234, "filterKey": -1},
	}

	AssertEqual(t, config.HandleEvent(valid_event), NoError)
	AssertEqual(t, config.HandleEvent(invalid_event), FailedFilter)

}

func TestMetricStorageInit(t *testing.T) {
	event_ids := []int{1}
	filter := NullFilter{}
	storage := NewNaiveStorage()
	config := MetricConfig{
		Id:         1,
		Name:       "testConfig",
		EventIds:   event_ids,
		KeyField:   "key",
		CountField: "",
		MetricType: CountMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create event
	event := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234},
	}

	config.HandleEvent(event)

	// Metric 1, for key 1234
	storage_result := storage.Get("1:1234")

	AssertEqual(t, storage_result.ErrCode, 0)
	AssertEqual(t, storage_result.Value, 1)
}

func TestMetricStateMaintained(t *testing.T) {
	event_ids := []int{1}
	filter := NullFilter{}
	storage := NewNaiveStorage()
	config := MetricConfig{
		Id:         1,
		Name:       "testConfig",
		EventIds:   event_ids,
		KeyField:   "key",
		CountField: "",
		MetricType: CountMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create event
	event := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234},
	}

	config.HandleEvent(event)
	storage_result_1 := storage.Get("1:1234")
	AssertEqual(t, storage_result_1.ErrCode, 0)
	AssertEqual(t, storage_result_1.Value, 1)

	config.HandleEvent(event)
	storage_result_2 := storage.Get("1:1234")
	AssertEqual(t, storage_result_2.ErrCode, 0)
	AssertEqual(t, storage_result_2.Value, 2)

}

func TestIncDecWithCountKey(t *testing.T) {
	event_ids := []int{1}
	filter := NullFilter{}
	storage := NewNaiveStorage()
	config := MetricConfig{
		Id:         1,
		Name:       "testConfig",
		EventIds:   event_ids,
		KeyField:   "key",
		CountField: "countKey",
		MetricType: CountMetricType,
		Filter:     filter,
		Storage:    storage,
	}

	// Create events
	event_1 := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234, "countKey": 5},
	}
	event_2 := Event{
		Id:   0,
		Data: map[string]interface{}{"key": 1234, "countKey": -2},
	}

	config.HandleEvent(event_1)
	storage_result_1 := storage.Get("1:1234")
	AssertEqual(t, storage_result_1.ErrCode, 0)
	AssertEqual(t, storage_result_1.Value, 5)

	config.HandleEvent(event_2)
	storage_result_2 := storage.Get("1:1234")
	AssertEqual(t, storage_result_2.ErrCode, 0)
	AssertEqual(t, storage_result_2.Value, 3)
}
