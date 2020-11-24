package configvalidator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
)

func TestConfigValidator(t *testing.T) {
	ct.AssertEqual(t, ValidateFilterString(`["null"`, map[string]int{}), false)
	ct.AssertEqual(t, ValidateFilterString(`["null"]`, map[string]int{}), true)
	ct.AssertEqual(t, ValidateFilterString(`["null", 1]`, map[string]int{}), false)
	ct.AssertEqual(t, ValidateFilterString(`["gt", "test", 1]`, map[string]int{}), false)
	ct.AssertEqual(t, ValidateFilterString(`["gt", "test", 1]`, map[string]int{"test": 1}), true)

	ct.AssertEqual(t, ValidateFilterString(`["all", ["null"]]`, map[string]int{}), true)
	ct.AssertEqual(t, ValidateFilterString(`["all", ["null", "null"]]`, map[string]int{}), false)
	ct.AssertEqual(t, ValidateFilterString(`["all", ["null"], ["null"]]`, map[string]int{}), true)
	ct.AssertEqual(t, ValidateFilterString(`["all", ["gt", "test", 1]]`, map[string]int{}), false)
	ct.AssertEqual(t, ValidateFilterString(`["all", ["gt", "test", 1]]`, map[string]int{"test": 1}), true)
}
