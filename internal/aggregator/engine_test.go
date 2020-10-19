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
	ecs := map[int]*eventConfig{1: &ec}
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
	mcs := map[int]*metricConfig{1: &mc}
	nsm := NewNSM(storage, true)
	nsm.SetNamespaceFromConfig("global", ecs, mcs)
	// Create engine
	engine := NewEngine(&nsm)

	// Handle a basic event
	result := engine.HandleRawEvent(re, "global")
	assert.Equal(t, nil, result)

	// Inspect the storage directly to check result
	// Metric 1, for key 1234

	sr := storage.Get("global:1:1234")
	ct.AssertEqual(t, sr.Err, nil)
	ct.AssertEqual(t, sr.Value, 1)
}

func TestNaiveE2E(t *testing.T) {
	storage := newNaiveStorage()
	E2ETest(t, storage)
}

func E2ETest(t *testing.T, storage AbstractStorage) {
	input, err := ioutil.ReadFile("../../config/aggregator_configs/global")
	if err != nil {
		panic(err)
	}
	nsm := NewNSM(storage, true)
	nsm.SetNamespaceFromData(input)
	engine := NewEngine(&nsm)

	// Handle a filtered event
	re1 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	res1 := engine.HandleRawEvent(re1, "global")
	sr1 := storage.Get("global:1:1234")

	ct.AssertEqual(t, res1, nil)
	ct.AssertEqual(t, sr1.Err, &StorageKeyNotFoundError{})
	ct.AssertEqual(t, sr1.Value, 0)

	// Handle a valid event
	re2 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 10, "test3": 1234, "test4": 1234}
	res2 := engine.HandleRawEvent(re2, "global")
	sr2 := storage.Get("global:1:1234")

	ct.AssertEqual(t, res2, nil)
	ct.AssertEqual(t, sr2.Err, nil)
	ct.AssertEqual(t, sr2.Value, 1234)

	// Handle another filtered event
	re3 := map[string]interface{}{"id": 1, "test1": 1234, "test2": 1, "test3": 1234, "test4": 1234}
	res3 := engine.HandleRawEvent(re3, "global")
	sr3 := storage.Get("global:1:1234")

	ct.AssertEqual(t, res3, nil)
	ct.AssertEqual(t, sr3.Err, nil)
	ct.AssertEqual(t, sr3.Value, 1234)
}

func TestNamespace(t *testing.T) {
	storage := newNaiveStorage()
	input1, err := ioutil.ReadFile("../../config/aggregator_configs/global")
	if err != nil {
		panic(err)
	}
	input2, err := ioutil.ReadFile("../../config/aggregator_configs/test")
	if err != nil {
		panic(err)
	}
	nsm := NewNSM(storage, true)
	nsm.SetNamespaceFromData(input1)
	nsm.SetNamespaceFromData(input2)

	engine := NewEngine(&nsm)

	// Handle a basic
	re1 := map[string]interface{}{"id": 1, "test1": 2, "test2": 2, "test3": 3, "test4": 4}

	res1 := engine.HandleRawEvent(re1, "global")
	sr1 := storage.Get("global:1:2")

	ct.AssertEqual(t, res1, nil)
	ct.AssertEqual(t, sr1.Err, nil)
	ct.AssertEqual(t, sr1.Value, 2)

	re2 := map[string]interface{}{"id": 1, "test1": 2, "test2": 2, "test3": 3, "test4": 4}
	res2 := engine.HandleRawEvent(re2, "test")
	sr2 := storage.Get("test:1:2")

	ct.AssertEqual(t, res2, nil)
	ct.AssertEqual(t, sr2.Err, nil)
	ct.AssertEqual(t, sr2.Value, 3)
}
