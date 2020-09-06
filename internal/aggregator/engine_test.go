package aggregator

import (
	"io/ioutil"
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
	"github.com/stretchr/testify/assert"
)

func TestEngine(t *testing.T) {
	// Create storage
	storage := newNaiveStorage()

	// Create event configs
	fields := map[string]fieldType{"key": intField}
	ec := eventConfig{
		Name:   "testEvent",
		ID:     1,
		Fields: fields,
	}
	ecs := []*eventConfig{&ec}
	eIDs := []int{1}

	// Create raw event
	re := map[string]interface{}{"id": 1, "key": 1234}

	// Create metric config
	mc := metricConfig{
		ID:         1,
		Name:       "testMetric",
		EventIds:   eIDs,
		KeyField:   "key",
		CountField: "",
		MetricType: countMetricType,
		Filter:     NullFilter{},
		Storage:    storage,
	}
	mcs := []*metricConfig{&mc}
	parser := newConfigParserFromConfigs(ecs, mcs, storage)

	// Create engine
	engine := newEngine(&parser)

	// Handle a basic event
	result := engine.HandleRawEvent(re, "")
	assert.Equal(t, success, result)

	// Inspect the storage directly to check result
	// Metric 1, for key 1234

	sr := storage.Get(":1:1234")
	ct.AssertEqual(t, sr.ErrCode, 0)
	ct.AssertEqual(t, sr.Value, 1)
}

func TestNaiveE2E(t *testing.T) {
	storage := newNaiveStorage()
	E2ETest(t, storage)
}

func E2ETest(t *testing.T, storage AbstractStorage) {
	input, _ := ioutil.ReadFile("../../config/example")
	parser := newConfigParserFromRaw(input, storage)
	engine := newEngine(&parser)

	// Handle a filtered event
	re1 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	res1 := engine.HandleRawEvent(re1, "")
	sr1 := storage.Get(":1:1234")

	ct.AssertEqual(t, res1, success)
	ct.AssertEqual(t, sr1.ErrCode, 1)
	ct.AssertEqual(t, sr1.Value, 0)

	// Handle a valid event
	re2 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 10, "test3": 1234, "test4": 1234}
	res2 := engine.HandleRawEvent(re2, "")
	sr2 := storage.Get(":1:1234")

	ct.AssertEqual(t, res2, success)
	ct.AssertEqual(t, sr2.ErrCode, 0)
	ct.AssertEqual(t, sr2.Value, 1234)

	// Handle another filtered event
	re3 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	res3 := engine.HandleRawEvent(re3, "")
	sr3 := storage.Get(":1:1234")

	ct.AssertEqual(t, res3, success)
	ct.AssertEqual(t, sr3.ErrCode, 0)
	ct.AssertEqual(t, sr3.Value, 1234)
}

func TestNamespace(t *testing.T) {
	storage := newNaiveStorage()
	input, _ := ioutil.ReadFile("../../config/example")
	parser := newConfigParserFromRaw(input, storage)
	engine := newEngine(&parser)

	// Handle a basic
	re1 := map[string]interface{}{"id": 1, "test1": 2, "test2": 2, "test3": 3, "test4": 4}

	res1 := engine.HandleRawEvent(re1, "")
	sr1 := storage.Get(":1:2")

	ct.AssertEqual(t, res1, success)
	ct.AssertEqual(t, sr1.ErrCode, 0)
	ct.AssertEqual(t, sr1.Value, 2)

	re2 := map[string]interface{}{"id": 1, "test1": 2, "test2": 2, "test3": 3, "test4": 4}
	res2 := engine.HandleRawEvent(re2, "test")
	sr2 := storage.Get("test:1:2")

	ct.AssertEqual(t, res2, success)
	ct.AssertEqual(t, sr2.ErrCode, 0)
	ct.AssertEqual(t, sr2.Value, 3)
}
