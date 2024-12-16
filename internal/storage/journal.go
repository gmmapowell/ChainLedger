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

func NewJournaller() Journaller {
	return &DummyJournaller{}
}
