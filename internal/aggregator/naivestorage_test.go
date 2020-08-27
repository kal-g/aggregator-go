package aggregator

import (
	"testing"

	ct "github.com/kal-g/aggregator-go/internal/common_test"
)

func TestNaiveStorage1(t *testing.T) {
	ns := newNaiveStorage()
	ns.Put("testKey", 1)
	ct.AssertEqual(t, ns.Get("testKey").Value, 1)
}

func TestNaiveStorage2(t *testing.T) {
	ns := newNaiveStorage()
	ns.Put("testKey", 1)
	ns.Put("testKey", 5)
	ct.AssertEqual(t, ns.Get("testKey").Value, 5)
}
