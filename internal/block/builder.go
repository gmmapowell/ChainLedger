package block

import (
	"fmt"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

const delay = 5 * time.Second
const pause = 1 * time.Second

type BlockBuilder interface {
	Start()
}

type SleepBlockBuilder struct {
	clock helpers.Clock
}

func (builder *SleepBlockBuilder) Start() {
	go builder.Run()
}

func (builder *SleepBlockBuilder) Run() {
	timer := builder.clock.After(delay)
	for {
		blocktime := <-timer
		timer = builder.clock.After(delay)
		runAt := <-builder.clock.After(pause)
		fmt.Printf("Building block ending %s at %s\n", blocktime.IsoTime(), runAt.IsoTime())
	}
}

func NewBlockBuilder(clock helpers.Clock, journal storage.Journaller) BlockBuilder {
	return &SleepBlockBuilder{clock: clock}
}
