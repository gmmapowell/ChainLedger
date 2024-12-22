package storage

import (
	"fmt"

	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Journaller interface {
	RecordTx(tx *records.StoredTransaction) error
	ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]records.StoredTransaction, error)
}

type DummyJournaller struct {
}

// RecordTx implements Journaller.
func (d *DummyJournaller) RecordTx(tx *records.StoredTransaction) error {
	fmt.Printf("Recording tx with id %v\n", tx.TxID)
	return nil
}

func (d DummyJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]records.StoredTransaction, error) {
	return nil, nil
}

type MemoryJournaller struct {
	name string
	txs  []*records.StoredTransaction
}

// RecordTx implements Journaller.
func (d *MemoryJournaller) RecordTx(tx *records.StoredTransaction) error {
	d.txs = append(d.txs, tx)
	fmt.Printf("%s recording tx with id %v, have %d\n", d.name, tx.TxID, len(d.txs))
	return nil
}

func (d MemoryJournaller) ReadTransactionsBetween(from types.Timestamp, upto types.Timestamp) ([]records.StoredTransaction, error) {
	return nil, nil
}

func NewJournaller(name string) Journaller {
	return &MemoryJournaller{name: name}
}
