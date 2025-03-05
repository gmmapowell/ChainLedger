package block

import (
	"crypto/rsa"
	"log"
	"net/url"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

const delay = 5 * time.Second
const pause = 1 * time.Second

type BlockBuilder interface {
	Start()
}

type SleepBlockBuilder struct {
	Name       *url.URL
	journaller storage.Journaller
	blocker    *Blocker
	clock      helpers.Clock
	control    types.PingBack
	senders    []helpers.BinarySender
}

func (builder *SleepBlockBuilder) Start() {
	go builder.Run()
}

func (builder *SleepBlockBuilder) Run() {
	blocktime := builder.clock.Time()
	timer := builder.clock.After(delay)
	lastBlock, err := builder.blocker.Build(blocktime, nil, nil)
	if err != nil {
		panic("error returned from building block 0")
	}
	builder.journaller.RecordBlock(lastBlock)
	lastBlock.MarshalAndSend(builder.senders)
	for {
		prev := blocktime
		select {
		case pingback := <-builder.control:
			log.Printf("%s asked to build final block and quit\n", builder.Name.String())
			builder.buildRecordAndSend(prev, builder.clock.Time(), lastBlock)
			pingback.Send()
			return
		case blocktime = <-timer:
			timer = builder.clock.After(delay)
			nowis := <-builder.clock.After(pause)
			// we are ready to build a block
			log.Printf("%s timer fired to build block: %s\n", builder.Name.String(), nowis.IsoTime())
			lastBlock = builder.buildRecordAndSend(prev, blocktime, lastBlock)
		}
	}
}

func (builder *SleepBlockBuilder) buildRecordAndSend(prevTime types.Timestamp, currTime types.Timestamp, lastBlock *records.Block) *records.Block {
	block := builder.buildBlock(prevTime, currTime, lastBlock)
	builder.journaller.RecordBlock(block)
	block.MarshalAndSend(builder.senders)
	return block
}

func (builder *SleepBlockBuilder) buildBlock(prev types.Timestamp, blocktime types.Timestamp, lastBlock *records.Block) *records.Block {
	txs, _ := builder.journaller.ReadTransactionsBetween(prev, blocktime)
	lastBlock, err := builder.blocker.Build(blocktime, lastBlock, txs)
	if err != nil {
		panic("error returned from building block")
	}
	return lastBlock
}

func NewBlockBuilder(clock helpers.Clock, journal storage.Journaller, url *url.URL, pk *rsa.PrivateKey, control types.PingBack, senders []helpers.BinarySender) BlockBuilder {
	hf := helpers.SHA512Factory{}
	sf := helpers.RSASigner{}
	blocker := NewBlocker(&hf, &sf, url, pk)
	return &SleepBlockBuilder{Name: url, clock: clock, journaller: journal, blocker: blocker, control: control, senders: senders}
}
