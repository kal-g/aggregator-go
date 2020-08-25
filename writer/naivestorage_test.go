package aggregator_test

import (
	"testing"

	. "github.com/kal-g/aggregator-go/writer"
)

func TestNaiveStorage1(t *testing.T) {
	ns := NewNaiveStorage()
	ns.Put("testKey", 1)
	AssertEqual(t, ns.Get("testKey").Value, 1)
}

func TestNaiveStorage2(t *testing.T) {
	ns := NewNaiveStorage()
	ns.Put("testKey", 1)
	ns.Put("testKey", 5)
	AssertEqual(t, ns.Get("testKey").Value, 5)
}
