package block

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
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
	journaller storage.Journaller
	blocker    *Blocker
	clock      helpers.Clock
}

func (builder *SleepBlockBuilder) Start() {
	go builder.Run()
}

func (builder *SleepBlockBuilder) Run() {
	blocktime := builder.clock.Time()
	timer := builder.clock.After(delay)
	lastBlock := builder.blocker.Build(blocktime, nil, nil)
	for {
		prev := blocktime
		blocktime = <-timer
		timer = builder.clock.After(delay)
		<-builder.clock.After(pause)
		txs, _ := builder.journaller.ReadTransactionsBetween(prev, blocktime)
		lastBlock = builder.blocker.Build(blocktime, lastBlock, txs)
	}
}

func NewBlockBuilder(clock helpers.Clock, journal storage.Journaller) BlockBuilder {
	url, _ := url.Parse("https://node1.com")
	pk, _ := rsa.GenerateKey(rand.Reader, 16)
	hf := helpers.SHA512Factory{}
	blocker := NewBlocker(&hf, url, pk)
	return &SleepBlockBuilder{clock: clock, journaller: journal, blocker: blocker}
}
