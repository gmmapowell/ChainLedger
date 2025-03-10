package helpers

import (
	"testing"
)

type ChanCollector struct {
	t         *testing.T
	collector chan any
}

// Fail implements Fatals.
func (cc *ChanCollector) Fail() {
	defer handleClosedChannel(cc)
	close(cc.collector)
}

// Fatalf implements Fatals.
func (cc *ChanCollector) Fatalf(format string, args ...any) {
	cc.t.Logf(format, args...)
	cc.Fail()
}

// Log implements Fatals.
func (cc *ChanCollector) Log(args ...any) {
	cc.t.Log(args...)
}

// Logf implements Fatals.
func (cc *ChanCollector) Logf(format string, args ...any) {
	cc.t.Logf(format, args...)
}

func (cc *ChanCollector) Send(obj any) {
	defer handleClosedChannel(cc)
	cc.collector <- obj
}

func (cc *ChanCollector) Recv() any {
	msg, ok := <-cc.collector
	if !ok {
		cc.t.FailNow()
	}
	return msg
}

func handleClosedChannel(cc Fatals) {
	if recover() != nil {
		cc.Logf("channel had already been closed")
	}
}

func NewChanCollector(t *testing.T, nc int) *ChanCollector {
	return &ChanCollector{t: t, collector: make(chan any, nc)}
}
