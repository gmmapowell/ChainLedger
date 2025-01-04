package storage_test

import (
	"fmt"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func TestWeCanAddAndRecoverAtTheSameTime(t *testing.T) {
	clock := helpers.ClockDoubleSameMinute("2024-12-25_03:00", "05.121", "07.282", "11.281", "19.202")

	finj := helpers.FaultInjectionLibrary(t)
	tj := storage.TestJournaller("journal", finj)
	journal := tj.(*storage.MemoryJournaller)
	completed := false
	go func() {
		for !completed {
			journal.RecordTx(storableTx(clock))
		}
	}()
	aw := finj.AllocatedWaiter()
	for !journal.HaveSome() || journal.NotAtCapacity() {
		aw.Release()
		aw = finj.AllocatedWaiter()
	}
	go func() {
		txs, _ := tj.ReadTransactionsBetween(clock.Times[0], clock.Times[3])
		fmt.Printf("%v\n", txs)
	}()
	rw := finj.AllocatedWaiter()
	aw.Release()
	/*aw = */ finj.AllocatedWaiter()
	rw.Release()
	// aw.Release()
}

func storableTx(clock helpers.Clock) *records.StoredTransaction {
	return &records.StoredTransaction{TxID: []byte("hello"), WhenReceived: clock.Time()}
}
