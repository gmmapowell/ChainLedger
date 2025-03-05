package storage_test

import (
	"log"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func TestWeCanAddAndRecoverAtTheSameTime(t *testing.T) {
	clock := helpers.ClockDoubleSameMinute("2024-12-25_03:00", "05.121", "07.282", "08.301", "08.402", "11.281", "14.010", "19.202")
	cc := helpers.NewChanCollector(t, 2)

	finj := helpers.FaultInjectionLibrary(cc)
	tj := storage.TestJournaller("journal", "myself", nil, finj)
	journal := tj.(*storage.MemoryJournaller)
	completed := false
	go func() {
		for !completed {
			journal.RecordTx(storableTx(clock))
		}
	}()
	aw := finj.AllocatedWaiter("journal-store-tx")
	for !journal.AtCapacityWithAtLeast(3) {
		aw.Release()
		aw = finj.AllocatedWaiter("journal-store-tx")
	}
	go func() {
		txs, _ := tj.ReadTransactionsBetween(clock.Times[0], clock.Times[6])
		log.Printf("%v\n", txs)
		txs, _ = tj.ReadTransactionsBetween(clock.Times[0], clock.Times[6])
		log.Printf("%v\n", txs)
		cc.Send(struct{}{})
	}()
	rw := finj.AllocatedWaiter("journal-read-txs")
	aw.Release()
	/*aw = */ finj.AllocatedWaiter("journal-store-tx")
	rw.Release()
	rw = finj.AllocatedWaiter("journal-read-txs")
	rw.Release()
	finj.AllowAll("journal-read-txs")
	cc.Recv()
	journal.Quit()
}

func storableTx(clock helpers.Clock) *records.StoredTransaction {
	return &records.StoredTransaction{TxID: []byte("hello"), WhenReceived: clock.Time()}
}
