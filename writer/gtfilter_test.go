package aggregator_test

import (
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestGreaterThanFilter(t *testing.T) {
	gt := GreaterThanFilter{
		Key:   "dataField1",
		Value: 0,
	}

	re1 := map[string]interface{}{"dataField1": 1}
	valid_event := Event{
		Id:   0,
		Data: re1,
	}

	re2 := map[string]interface{}{"dataField1": -1}
	invalid_event := Event{
		Id:   0,
		Data: re2,
	}

	AssertEqual(t, gt.IsValid(valid_event), true)
	AssertEqual(t, gt.IsValid(invalid_event), false)
}
