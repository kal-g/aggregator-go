package aggregator_test

import (
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestEvent(t *testing.T) {
	re1 := map[string]interface{}{"dataField1": 1, "dataField2": "test", "dataField3": 3}
	e := Event{
		Id:   0,
		Data: re1,
	}

	AssertEqual(t, e.GetDataField("dataField1"), 1)
	AssertEqual(t, e.GetDataField("dataField2"), "test")
	AssertEqual(t, e.GetDataField("dataField3"), 3)
}
