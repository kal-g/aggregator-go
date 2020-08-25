package aggregator_test

import (
	"io/ioutil"
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
	// Create storage
	storage := NewNaiveStorage()

	// Create event configs
	fields := map[string]FieldType{"key": IntField}
	event_config := EventConfig{
		Name:   "testEvent",
		Id:     1,
		Fields: fields,
	}
	event_configs := []EventConfig{event_config}
	event_ids := []int{1}

	// Create raw event
	raw_event := map[string]interface{}{"id": 1, "key": 1234}

	// Create metric config
	metric_config := MetricConfig{
		Id:         1,
		Name:       "testMetric",
		EventIds:   event_ids,
		KeyField:   "key",
		CountField: "",
		MetricType: CountMetricType,
		Filter:     NullFilter{},
		Storage:    storage,
	}
	metric_configs := []MetricConfig{metric_config}
	parser := NewConfigParserFromConfigs(event_configs, metric_configs, storage)

	// Create engine
	engine := NewEngine(&parser)

	// Handle a basic event
	result := engine.HandleRawEvent(raw_event)
	assert.Equal(t, Success, result)

	// Inspect the storage directly to check result
	// Metric 1, for key 1234

	storage_result := storage.Get("1:1234")
	AssertEqual(t, storage_result.ErrCode, 0)
	AssertEqual(t, storage_result.Value, 1)
}

func TestNaiveE2E(t *testing.T) {
	storage := NewNaiveStorage()
	E2ETest(t, storage)
}

func E2ETest(t *testing.T, storage AbstractStorage) {
	input, _ := ioutil.ReadFile("../tools/config/example")
	parser := NewConfigParserFromRaw(input, storage)
	engine := NewEngine(&parser)

	// Handle a filtered event
	re_1 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	result_1 := engine.HandleRawEvent(re_1)
	storage_result_1 := storage.Get("1:1234")

	AssertEqual(t, result_1, Success)
	AssertEqual(t, storage_result_1.ErrCode, 1)
	AssertEqual(t, storage_result_1.Value, 0)

	// Handle a valid event
	re_2 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 10, "test3": 1234, "test4": 1234}
	result_2 := engine.HandleRawEvent(re_2)
	storage_result_2 := storage.Get("1:1234")

	AssertEqual(t, result_2, Success)
	AssertEqual(t, storage_result_2.ErrCode, 0)
	AssertEqual(t, storage_result_2.Value, 10)

	// Handle another filtered event
	re_3 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	result_3 := engine.HandleRawEvent(re_3)
	storage_result_3 := storage.Get("1:1234")

	AssertEqual(t, result_3, Success)
	AssertEqual(t, storage_result_3.ErrCode, 0)
	AssertEqual(t, storage_result_3.Value, 10)
}
