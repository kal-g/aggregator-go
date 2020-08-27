package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
)

func TestEvent(t *testing.T) {
	re1 := map[string]interface{}{"dataField1": 1, "dataField2": "test", "dataField3": 3}
	e := event{
		ID:   0,
		Data: re1,
	}

	ct.AssertEqual(t, e.GetDataField("dataField1"), 1)
	ct.AssertEqual(t, e.GetDataField("dataField2"), "test")
	ct.AssertEqual(t, e.GetDataField("dataField3"), 3)
}
