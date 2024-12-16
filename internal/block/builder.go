package block

import (
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type BlockBuilder interface {
	Start()
}

type SleepBlockBuilder struct {
}

func (builder *SleepBlockBuilder) Start() {

}

func NewBlockBuilder(clock helpers.Clock, journal storage.Journaller) BlockBuilder {
	return &SleepBlockBuilder{}
}
