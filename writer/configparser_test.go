package aggregator

import (
	"fmt"
	"io/ioutil"
	"testing"

	ct "github.com/kal-g/aggregator-go/common_test"
)

func TestConfigParser(t *testing.T) {
	input, err := ioutil.ReadFile("../tools/config/example")
	ct.AssertEqual(t, err, nil)
	cp := newConfigParserFromRaw(input, nil)
	cpStr := fmt.Sprintf("%+v", cp)
	eventConfigs := "{EventConfigs:[{Name:testEvent ID:1 Fields:map[test1:2 test2:2 test3:2 test4:2]}] "
	metricConfigs := "MetricConfigs:[{ID:1 Name:testMetric EventIds:[1] KeyField:test1 CountField:test1 MetricType:0 Namespace: Filter:{Filters:[{Key:test1 Value:1} {Key:test2 Value:1}]} Storage:<nil>} {ID:1 Name:testMetric EventIds:[1] KeyField:test1 CountField:test3 MetricType:0 Namespace:test Filter:{Filters:[{Key:test1 Value:1} {Key:test2 Value:1}]} Storage:<nil>}] storage:<nil>}"
	ct.AssertEqual(t, cpStr, eventConfigs+metricConfigs)
}
