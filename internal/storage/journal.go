package storage

import (
	"fmt"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Journaller interface {
	RecordTx(tx *records.StoredTransaction) error
	ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error)
	Quit() error
}

type DummyJournaller struct {
}

// RecordTx implements Journaller.
func (d *DummyJournaller) RecordTx(tx *records.StoredTransaction) error {
	fmt.Printf("Recording tx with id %v\n", tx.TxID)
	return nil
}

func (d DummyJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error) {
	return nil, nil
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

func (d MemoryJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error) {
	messageMe := make(chan []*records.StoredTransaction)
	d.finj.NextWaiter("journal-read-txs")
	d.tothread <- JournalRetrieveCommand{From: from, Upto: upto, ResultChan: messageMe}
	ret := <-messageMe
	return ret, nil
}

func (d *MemoryJournaller) Quit() error {
	return nil
}

func (d *MemoryJournaller) AtCapacityWithAtLeast(n int) bool {
	messageMe := make(chan bool)
	d.tothread <- JournalCheckCapacityCommand{AtLeast: n, ResultChan: messageMe}
	return <-messageMe
}

func NewJournaller(name string) Journaller {
	return TestJournaller(name, helpers.IgnoreFaultInjection())
}

func TestJournaller(name string, finj helpers.FaultInjection) Journaller {
	ret := MemoryJournaller{name: name, finj: finj}
	ret.tothread = LaunchJournalThread(name, finj)
	return &ret
}
