package storage_test

import (
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func TestWeCanAddAndRecoverAtTheSameTime(t *testing.T) {
	finj := helpers.FaultInjectionLibrary(t)
	tj := storage.TestJournaller("journal", finj)
	journal := tj.(*storage.MemoryJournaller)
	completed := false
	go func() {
		for !completed {
			journal.RecordTx(storableTx())
		}
	}()
	aw := finj.AllocatedWaiter()
	for !journal.HaveSome() || journal.NotAtCapacity() {
		aw.Release()
		aw = finj.AllocatedWaiter()
	}

}

func storableTx() *records.StoredTransaction {
	return &records.StoredTransaction{TxID: []byte("hello")}
}
