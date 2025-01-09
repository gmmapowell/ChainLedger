package storage_test

import (
	"fmt"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func TestWeCanAddAndRecoverAtTheSameTime(t *testing.T) {
	clock := helpers.ClockDoubleSameMinute("2024-12-25_03:00", "05.121", "07.282", "08.301", "08.402", "11.281", "14.010", "19.202")

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
	for !journal.HaveAtLeast(3) || journal.NotAtCapacity() {
		aw.Release()
		aw = finj.AllocatedWaiter()
	}
	waitAll := make(chan struct{})
	go func() {
		txs, _ := tj.ReadTransactionsBetween(clock.Times[0], clock.Times[6])
		fmt.Printf("%v\n", txs)
		txs, _ = tj.ReadTransactionsBetween(clock.Times[0], clock.Times[6])
		fmt.Printf("%v\n", txs)
		waitAll <- struct{}{}
	}()
	rw := finj.AllocatedWaiter()
	aw.Release()
	/*aw = */ finj.AllocatedWaiter()
	rw.Release()
	finj.JustRun()
	<-waitAll
}

func storableTx(clock helpers.Clock) *records.StoredTransaction {
	return &records.StoredTransaction{TxID: []byte("hello"), WhenReceived: clock.Time()}
}
