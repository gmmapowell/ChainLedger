package loom

import (
	"log"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type LoomThread interface {
	Start()
}

type IntervalLoomThread struct {
	clock     helpers.Clock
	myjournal storage.Journaller
	interval  int
	control   <-chan string
	loom      *Loom
}

func (t *IntervalLoomThread) Start() {
	log.Printf("starting loom thread for %s\n", t.loom.Name())
	go t.Run()
}

func (t *IntervalLoomThread) Run() {
	delay := time.Duration(t.interval/3) * time.Millisecond
	timer := t.clock.After(delay)
	var prev *records.Weave

	for {
		select {
		case <-t.control:
			log.Printf("%s weaver asked to quit\n", t.loom.Name())
			return
		case weaveBefore := <-timer:
			weaveBefore = weaveBefore.RoundTime(t.interval)
			if !t.myjournal.HasWeaveAt(weaveBefore) {
				weave := t.loom.WeaveAt(weaveBefore, prev)
				if weave != nil {
					t.myjournal.StoreWeave(weave)
					log.Printf("%s wove at %v: %s\n", t.loom.Name(), weaveBefore, weave.ID.String())
					// weave.LogMe(t.loom.Name())
					prev = weave
				} else {
					log.Printf("%s could not weave at %v\n", t.loom.Name(), weaveBefore)
				}
			}
		}
		timer = t.clock.After(delay)
	}
}

func NewLoomThread(clock helpers.Clock, myname string, control <-chan string, interval int, loom *Loom, myjournal storage.Journaller) LoomThread {
	return &IntervalLoomThread{clock: clock, control: control, loom: loom, myjournal: myjournal, interval: interval}
}
