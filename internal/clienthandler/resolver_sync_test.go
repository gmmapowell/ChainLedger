package clienthandler_test

import (
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

func TestThatTwoThreadsCanSignDifferentFieldsAtTheSameTime(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	cc := helpers.NewChanCollector(t, 2)
	setup(cc, clock)
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
	s1 := cc.Recv()
	s2 := cc.Recv()
	if s1 != s2 {
		t.Fatalf("The two transactions were not the same")
	}
	tx1 := s1.(records.StoredTransaction)
	if tx1.Signatories[0].Signature == nil {
		t.Fatalf("the first signature is missing")
	}
	if tx1.Signatories[1].Signature == nil {
		t.Fatalf("the second signature is missing")
	}
}
