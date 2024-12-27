package storage_test

import (
	"testing"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

func TestTwoThreadsCannotBeInCriticalZoneAtOnce(t *testing.T) {
	finj := helpers.FaultInjectionLibrary(t)
	mps := storage.TestMemoryPendingStorage(finj)
	results := make(chan *api.Transaction, 2)
	go func() {
		tx1, _ := api.NewTransaction("https://hello.com", types.Hash("hello"))
		results <- tx1
		sx := mps.PendingTx(tx1)
		results <- sx
	}()
	w1 := finj.AllocatedWaiter()

	go func() {
		tx2, _ := api.NewTransaction("https://hello.com", types.Hash("hello"))
		results <- tx2
		rx := mps.PendingTx(tx2)
		results <- rx
	}()
	w2 := finj.AllocatedWaiterOrNil(50 * time.Millisecond)

	if w2 != nil {
		t.Fatalf("second waiter allocated before first released")
	}
	w1.Release()

	w2 = finj.AllocatedWaiter()
	if w2 == nil {
		t.Fatalf("second waiter could not be allocated after first released")
	}
	w2.Release()

	tx1 := <-results
	tx2 := <-results
	rx1 := <-results
	rx2 := <-results

	if rx1 != nil {
		t.Fatalf("the first call to PendingTx should return nil, not %v\n", rx1)
	}
	if tx1 != rx2 {
		t.Fatalf("we did not get back the same transaction: %v %v\n", tx1, rx2)
	}
	if tx2 == rx2 {
		t.Fatalf("we received the same second tx: %v %v\n", tx2, rx2)
	}
}
