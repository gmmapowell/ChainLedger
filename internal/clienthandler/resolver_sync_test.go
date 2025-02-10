package clienthandler_test

import (
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

func TestThatTwoThreadsCanSignDifferentFieldsAtTheSameTime(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	cc := helpers.NewChanCollector(t, 2)
	setup(cc, "https://test.com/node", clock, true)

	h1 := hasher.AddMock("fred")
	h1.AcceptAnything()

	signer.Expect(types.Signature("tx-sig"), nodeKey, types.Hash("fred"))

	go func() {
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
		stx, _ := r.ResolveTx(tx)
		cc.Send(stx)
	}()
	go func() {
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
		stx, _ := r.ResolveTx(tx)
		cc.Send(stx)
	}()

	// Now wait for both of them to get to the critical section
	w1 := finj.AllocatedWaiter("resolve-tx")
	w2 := finj.AllocatedWaiter("resolve-tx")

	// Then we can release both of them
	w1.Release()
	w2.Release()

	s1 := cc.Recv()
	s2 := cc.Recv()
	tx1 := s1.(*records.StoredTransaction)
	tx2 := s2.(*records.StoredTransaction)
	if tx1 == nil && tx2 == nil {
		t.Fatalf("both transactions were nil")
	}
	if tx1 != nil && tx2 != nil {
		t.Fatalf("both transactions were NOT nil: %v %v", tx1, tx2)
	}
	if tx1 == nil {
		tx1 = tx2
	}
	if tx1.Signatories[0].Signature == nil {
		t.Fatalf("the first signature is missing")
	}
	if tx1.Signatories[1].Signature == nil {
		t.Fatalf("the second signature is missing")
	}
}
