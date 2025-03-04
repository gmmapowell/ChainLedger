package loom

import (
	"time"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type LoomThread interface {
	Start()
}

type IntervalLoomThread struct {
	clock     helpers.Clock
	myjournal storage.Journaller
	interval  int
	loom      *Loom
}

func (t *IntervalLoomThread) Start() {
	go t.Run()
}

func (t *IntervalLoomThread) Run() {
	delay := time.Duration(t.interval/3) * time.Millisecond
	timer := t.clock.After(delay)

	for {
		select {
		case weaveBefore := <-timer:
			weaveBefore = weaveBefore.RoundTime(t.interval)
			if !t.myjournal.HasWeaveAt(weaveBefore) {
				weave := t.loom.WeaveAt(weaveBefore)
				t.myjournal.StoreWeave(weave)
			}
		}
		timer = t.clock.After(delay)

	}
}

func NewLoomThread(clock helpers.Clock, myname string, interval int, myjournal storage.Journaller) LoomThread {
	loom := &Loom{myname: myname}
	return &IntervalLoomThread{clock: clock, loom: loom, myjournal: myjournal, interval: interval}
}
