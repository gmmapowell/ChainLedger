package storage

import (
	"fmt"

	"github.com/gmmapowell/ChainLedger/internal/records"
)

type Journaller interface {
	RecordTx(tx *records.StoredTransaction) error
}

type DummyJournaller struct {
}

// RecordTx implements Journaller.
func (d *DummyJournaller) RecordTx(tx *records.StoredTransaction) error {
	fmt.Printf("Recording tx with id %v\n", tx.TxID)
	return nil
}

func NewJournaller() Journaller {
	return &DummyJournaller{}
}
