package storage

import (
	"fmt"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Journaller interface {
	RecordTx(tx *records.StoredTransaction) error
	RecordBlock(block *records.Block) error
	HasBlock(id types.Hash) bool
	CheckTxs(ids []types.Hash) []types.Hash
	HasWeaveAt(when types.Timestamp) bool
	StoreWeave(weave *records.Weave)
	RecordWeaveSignature(when types.Timestamp, id types.Hash, signature types.Signature) error
	ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error)
	LatestBlockBy(when types.Timestamp) types.Hash
	Quit() error
}

type DummyJournaller struct {
}

// RecordTx implements Journaller.
func (d *DummyJournaller) RecordTx(tx *records.StoredTransaction) error {
	fmt.Printf("Recording tx with id %v\n", tx.TxID)
	return nil
}

func (d *DummyJournaller) RecordBlock(block *records.Block) error {
	return nil
}

func (d DummyJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error) {
	return nil, nil
}

func (d *DummyJournaller) HasBlock(id types.Hash) bool {
	return true
}

func (d *DummyJournaller) CheckTxs(ids []types.Hash) []types.Hash {
	return nil
}

func (d *DummyJournaller) HasWeaveAt(when types.Timestamp) bool {
	return false
}

func (d *DummyJournaller) StoreWeave(weave *records.Weave) {
}

func (d *DummyJournaller) RecordWeaveSignature(when types.Timestamp, id types.Hash, signature types.Signature) error {
	return nil
}

func (d *DummyJournaller) LatestBlockBy(when types.Timestamp) types.Hash {
	return types.Hash{}
}

func (d *DummyJournaller) Quit() error {
	return nil
}

func NewDummyJournaller() Journaller {
	return &DummyJournaller{}
}

type MemoryJournaller struct {
	name     string
	tothread chan<- JournalCommand
	finj     helpers.FaultInjection
}

// RecordTx implements Journaller.
func (d *MemoryJournaller) RecordTx(tx *records.StoredTransaction) error {
	d.finj.NextWaiter("journal-store-tx")
	d.tothread <- JournalStoreCommand{Tx: tx}
	return nil
}

func (d *MemoryJournaller) RecordBlock(block *records.Block) error {
	d.tothread <- JournalBlockCommand{Block: block}
	return nil
}

func (d MemoryJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error) {
	messageMe := make(chan []*records.StoredTransaction)
	d.finj.NextWaiter("journal-read-txs")
	d.tothread <- JournalRetrieveCommand{From: from, Upto: upto, ResultChan: messageMe}
	ret := <-messageMe
	return ret, nil
}

func (d *MemoryJournaller) HasBlock(id types.Hash) bool {
	messageMe := make(chan bool)
	d.finj.NextWaiter("journal-has-block")
	d.tothread <- JournalHasBlockCommand{ID: id, ResultChan: messageMe}
	ret := <-messageMe
	return ret
}

func (d *MemoryJournaller) CheckTxs(ids []types.Hash) []types.Hash {
	messageMe := make(chan []types.Hash)
	d.finj.NextWaiter("journal-check-txs")
	d.tothread <- JournalCheckTxsCommand{IDs: ids, ResultChan: messageMe}
	ret := <-messageMe
	return ret
}

func (d *MemoryJournaller) HasWeaveAt(when types.Timestamp) bool {
	messageMe := make(chan bool)
	d.finj.NextWaiter("journal-has-weave-at")
	d.tothread <- JournalHasWeaveAtCommand{When: when, ResultChan: messageMe}
	ret := <-messageMe
	return ret
}

func (d *MemoryJournaller) StoreWeave(weave *records.Weave) {
	d.finj.NextWaiter("journal-store-weave")
	d.tothread <- JournalStoreWeaveCommand{Weave: weave}
}

func (d *MemoryJournaller) RecordWeaveSignature(when types.Timestamp, id types.Hash, signature types.Signature) error {
	d.finj.NextWaiter("journal-record-weave-signature")
	d.tothread <- JournalRecordWeaveSignatureCommand{When: when, ID: id, Signature: signature}
	return nil
}

func (d *MemoryJournaller) LatestBlockBy(when types.Timestamp) types.Hash {
	messageMe := make(chan types.Hash)
	d.finj.NextWaiter("journal-latest-block-by")
	d.tothread <- JournalLatestBlockByCommand{Before: when, ResultChan: messageMe}
	ret := <-messageMe
	return ret
}

func (d *MemoryJournaller) Quit() error {
	donech := make(chan struct{})
	d.tothread <- JournalDoneCommand{NotifyMe: donech}
	<-donech
	return nil
}

func (d *MemoryJournaller) AtCapacityWithAtLeast(n int) bool {
	messageMe := make(chan bool)
	d.tothread <- JournalCheckCapacityCommand{AtLeast: n, ResultChan: messageMe}
	return <-messageMe
}

func NewJournaller(forNode string, onNode string, consolidator *WeaveConsolidator) Journaller {
	return TestJournaller(forNode, onNode, consolidator, helpers.IgnoreFaultInjection())
}

func TestJournaller(forNode string, onNode string, consolidator *WeaveConsolidator, finj helpers.FaultInjection) Journaller {
	ret := MemoryJournaller{name: forNode, finj: finj}
	ret.tothread = LaunchJournalThread(forNode, onNode, consolidator, finj)
	return &ret
}
