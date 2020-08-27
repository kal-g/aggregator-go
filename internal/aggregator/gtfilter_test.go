package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
)

func TestGreaterThanFilter(t *testing.T) {
	gt := greaterThanFilter{
		Key:   "dataField1",
		Value: 0,
	}

	re1 := map[string]interface{}{"dataField1": 1}
	validEvent := event{
		ID:   0,
		Data: re1,
	}

	re2 := map[string]interface{}{"dataField1": -1}
	invalidEvent := event{
		ID:   0,
		Data: re2,
	}

	ct.AssertEqual(t, gt.IsValid(validEvent), true)
	ct.AssertEqual(t, gt.IsValid(invalidEvent), false)
}
