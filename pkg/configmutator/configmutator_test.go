package configmutator

import (
	"io/ioutil"
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
	"github.com/stretchr/testify/assert"
)

func TestConfigMutator(t *testing.T) {
	input, err := ioutil.ReadFile("../../config/aggregator_configs/global")
	if err != nil {
		panic(err)
	}
	cm := NewConfigMutator(string(input))

	ct.AssertEqual(t, cm.C.Namespace, "global")

	ct.AssertEqual(t, cm.C.Metrics[1].ID, 1)
	ct.AssertEqual(t, cm.C.Metrics[1].Name, "testMetric")
	ct.AssertEqual(t, cm.C.Metrics[1].EventIDs, []int{1})
	ct.AssertEqual(t, cm.C.Metrics[1].KeyField, "test1")
	ct.AssertEqual(t, cm.C.Metrics[1].Type, "count")
	ct.AssertEqual(t, cm.C.Metrics[1].Filter, []interface{}{
		"all",
		[]interface{}{"gt", "test1", float64(1)},
		[]interface{}{"gt", "test2", float64(1)},
	})
	ct.AssertEqual(t, cm.C.Metrics[1].CountField, "test1")

	ct.AssertEqual(t, cm.C.Events[1].ID, 1)
	ct.AssertEqual(t, cm.C.Events[1].Name, "testEvent")
	ct.AssertEqual(t, cm.C.Events[1].Fields, map[string]int{
		"test1": 1,
		"test2": 1,
		"test3": 1,
		"test4": 1,
	})

	// Add New Event
	cm.AddNewEvent("testEvent2")
	ct.AssertEqual(t, cm.C.Events[2].ID, 2)
	ct.AssertEqual(t, cm.C.Events[2].Name, "testEvent2")
	ct.AssertEqual(t, cm.C.Events[2].Fields, map[string]int{})

	// Add Event Field
	err = cm.AddEventField(3, "test1", 0)
	ct.AssertEqual(t, err, &EventNotFoundError{})
	err = cm.AddEventField(2, "test1", 0)
	ct.AssertEqual(t, err, &FieldTypeConflict{})
	err = cm.AddEventField(1, "test1", 1)
	ct.AssertEqual(t, err, &FieldAlreadyExists{})
	err = cm.AddEventField(2, "test1", 1)
	ct.AssertEqual(t, err, nil)

	// Add New Metric
	err = cm.AddNewMetric("testMetric2", "test", "test")
	ct.AssertEqual(t, err, &InvalidKeyField{})
	err = cm.AddNewMetric("testMetric2", "test1", "test")
	ct.AssertEqual(t, err, &InvalidCountField{})
	err = cm.AddNewMetric("testMetric2", "test1", "test1")
	ct.AssertEqual(t, err, nil)

	ids := cm.GetNewEventIDsForMetric(1)
	ct.AssertEqual(t, ids, []int{2})

	ids = cm.GetNewEventIDsForMetric(2)
	assert.ElementsMatch(t, ids, []int{1, 2})

}
