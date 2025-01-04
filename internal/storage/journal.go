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

type MemoryJournaller struct {
	name string
	txs  []*records.StoredTransaction
	finj helpers.FaultInjection
}

// RecordTx implements Journaller.
func (d *MemoryJournaller) RecordTx(tx *records.StoredTransaction) error {
	d.finj.NextWaiter()
	d.txs = append(d.txs, tx)
	fmt.Printf("%s recording tx with id %v, have %d at %p\n", d.name, tx.TxID, len(d.txs), d.txs)
	return nil
}

func (d MemoryJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]*records.StoredTransaction, error) {
	var ret []*records.StoredTransaction
	for _, tx := range d.txs {
		fmt.Printf("before waiting txs = %p\n", d.txs)
		d.finj.NextWaiter()
		fmt.Printf("after waiting txs = %p\n", d.txs)
		if tx.WhenReceived >= from && tx.WhenReceived < upto {
			ret = append(ret, tx)
		}
	}
	return ret, nil
}

func (d *MemoryJournaller) HaveSome() bool {
	fmt.Printf("len = %d\n", len(d.txs))
	return len(d.txs) > 0
}

func (d *MemoryJournaller) NotAtCapacity() bool {
	fmt.Printf("cap = %d len = %d\n", cap(d.txs), len(d.txs))
	return cap(d.txs) < len(d.txs)
}

func NewJournaller(name string) Journaller {
	return TestJournaller(name, helpers.IgnoreFaultInjection())
}

func TestJournaller(name string, finj helpers.FaultInjection) Journaller {
	return &MemoryJournaller{name: name, finj: finj}
}
