package aggregator_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestConfigParser(t *testing.T) {
	input, err := ioutil.ReadFile("../tools/config/example")
	AssertEqual(t, err, nil)
	cp := NewConfigParserFromRaw(input, nil)
	cp_str := fmt.Sprintf("%+v\n", cp)
	event_configs := "{EventConfigs:[{Name:testEvent Id:1 Fields:map[test1:2 test2:2 test3:2 test4:2]}] "
	metric_configs := "MetricConfigs:[{Id:1 Name:testMetric EventIds:[1] KeyField:test1 CountField:test2 MetricType:0 Filter:{Filters:[{Key:test1 Value:1} {Key:test2 Value:1}]} Storage:<nil>}] storage:<nil>}\n"
	AssertEqual(t, cp_str, event_configs+metric_configs)
}
