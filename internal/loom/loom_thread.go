package loom

import "github.com/gmmapowell/ChainLedger/internal/helpers"

type LoomThread interface {
	Start()
}

type IntervalLoomThread struct {
	clock helpers.Clock
	loom  *Loom
}

func (t *IntervalLoomThread) Start() {
	go t.Run()
}

func (t *IntervalLoomThread) Run() {
	t.loom.WeaveAt(t.clock.Time())
}
