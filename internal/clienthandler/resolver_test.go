package clienthandler_test

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

var repo client.ClientRepository
var s storage.PendingStorage
var r clienthandler.Resolver

func setup(clock helpers.Clock) {
	repo, _ = client.MakeMemoryRepo()
	s = storage.NewMemoryPendingStorage()
	r = clienthandler.NewResolver(clock, s)
}

func maketx(link string, hash string, userkeys ...any) *api.Transaction {
	tx, _ := api.NewTransaction(link, types.Hash([]byte(hash)))
	var ui *url.URL
	for _, v := range userkeys {
		if vs, ok := v.(string); ok {
			ui, _ = url.Parse(vs)
			tx.Signer(ui)
		} else if vb, ok := v.(bool); ok && vb {
			pk, _ := repo.PrivateKey(ui)
			tx.Sign(ui, pk)
		}
	}
	return tx
}

func TestANewTransactionMayBeStoredButReturnsNothing(t *testing.T) {
	setup(nil)
	tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	stx, err := r.ResolveTx(tx)
	checkNotReturned(t, stx, err)
}

func TestTwoCopiesOfTheTransactionAreEnoughToContinue(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(&clock)
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
		stx, err := r.ResolveTx(tx)
		checkNotReturned(t, stx, err)
	}
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
		stx, _ := r.ResolveTx(tx)
		if stx == nil {
			t.Fatalf("a stored transaction was not returned after both parties had submitted a signed copy")
		}
	}
}

func TestTwoIndependentTxsCanExistAtOnce(t *testing.T) {
	setup(nil)
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
		stx, err := r.ResolveTx(tx)
		checkNotReturned(t, stx, err)
	}
	{
		tx := maketx("https://test.com/msg2", "hash4", "https://user1.com/", "https://user2.com/", true)
		stx, err := r.ResolveTx(tx)
		checkNotReturned(t, stx, err)
	}
}

func TestTheReturnedTxHasAllTheFields(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(&clock)
	tx1 := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	r.ResolveTx(tx1)
	tx2 := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
	stx, _ := r.ResolveTx(tx2)
	if stx == nil {
		t.Fatalf("a stored transaction was not returned after both parties had submitted a signed copy")
	}
	if stx.ContentLink == nil {
		t.Fatalf("the stored transaction did not have the ContentLink")
	}
	if *stx.ContentLink != *tx1.ContentLink {
		t.Fatalf("the stored transaction ContentLink did not match")
	}
	if !bytes.Equal(stx.ContentHash, tx1.ContentHash) {
		t.Fatalf("the stored transaction ContentHash did not match")
	}
	if len(stx.Signatories) != len(tx1.Signatories) {
		t.Fatalf("the stored transaction did not have the correct number of signatories (%d not %d)", len(stx.Signatories), len(tx1.Signatories))
	}
	checkSignature(t, 0, stx.Signatories, tx1.Signatories)
	checkSignature(t, 1, stx.Signatories, tx2.Signatories)
}

func TestTheReturnedTxHasATimestamp(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(&clock)
	tx1 := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	r.ResolveTx(tx1)
	tx2 := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
	stx, _ := r.ResolveTx(tx2)
	if stx == nil {
		t.Fatalf("a stored transaction was not returned after both parties had submitted a signed copy")
	}
	if stx.WhenReceived != clock.Times[0] {
		t.Fatalf("the stored transaction was received at %s not %s", stx.WhenReceived.IsoTime(), clock.Times[0].IsoTime())
	}
	if stx.TxID == nil {
		t.Fatalf("the stored transaction did not have a TxID")
	}
}

func checkNotReturned(t *testing.T, stx *records.StoredTransaction, err error) {
	if stx != nil {
		t.Fatalf("a stored transaction was returned when the message was not fully signed")
	}
	if err != nil {
		t.Fatalf("ResolveTx returned an error: %v\n", err)
	}
}

func checkSignature(t *testing.T, which int, blockA []*types.Signatory, blockB []*types.Signatory) {
	sigA := blockA[which]
	sigB := blockB[which]
	if sigA.Signer.String() != sigB.Signer.String() {
		t.Fatalf("Signer for %d did not match: %s not %s", which, sigA.Signer.String(), sigB.Signer.String())
	}
	if !bytes.Equal(*sigA.Signature, *sigB.Signature) {
		t.Fatalf("Signature for %d did not match: %x not %x", which, sigA.Signature, sigB.Signature)
	}
}
