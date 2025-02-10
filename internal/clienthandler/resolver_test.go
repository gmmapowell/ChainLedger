package clienthandler_test

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
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

var hasher *helpers.MockHasherFactory
var signer *helpers.MockSigner
var repo client.MemoryClientRepository
var nodeKey *rsa.PrivateKey
var s storage.PendingStorage
var r clienthandler.Resolver
var finj helpers.FaultInjection

func setup(t helpers.Fatals, nodeName string, clock helpers.Clock, wantFi bool) {
	nodeURL, _ := url.Parse(nodeName)
	if wantFi {
		finj = helpers.FaultInjectionLibrary(t)
	} else {
		finj = helpers.IgnoreFaultInjection()
	}
	hasher = helpers.NewMockHasherFactory(t)
	signer = helpers.NewMockSigner(t, nodeURL)
	repo, _ = client.MakeMemoryRepo()
	repo.NewUser("https://user1.com/")
	repo.NewUser("https://user2.com/")
	nodeKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	s = storage.NewMemoryPendingStorage()
	r = clienthandler.TestResolver(finj, clock, hasher, signer, nodeKey, s)
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
	setup(t, "https://test.com/node", nil, false)
	tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	stx, err := r.ResolveTx(tx)
	checkNotReturned(t, stx, err)
}

func TestTwoCopiesOfTheTransactionAreEnoughToContinue(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(t, "https://test.com/node", clock, false)
	h1 := hasher.AddMock("fred")
	h1.ExpectTimestamp(clock.Times[0])
	h1.ExpectString("https://test.com/msg1\n")
	h1.ExpectString("hash")
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
		h1.ExpectString(("https://user1.com/\n"))
		h1.ExpectSignature(tx.Signatories[0].Signature)
		stx, err := r.ResolveTx(tx)
		checkNotReturned(t, stx, err)
	}
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
		h1.ExpectString(("https://user2.com/\n"))
		h1.ExpectSignature(tx.Signatories[1].Signature)
		signer.Expect(types.Signature("tx-sig"), nodeKey, types.Hash("fred"))
		stx, err := r.ResolveTx(tx)
		if err != nil {
			t.Fatalf("error on resolution: %v\n", err)
		} else if stx == nil {
			t.Fatalf("a stored transaction was not returned after both parties had submitted a signed copy")
		}
	}
}

func TestTwoIndependentTxsCanExistAtOnce(t *testing.T) {
	setup(t, "https://test.com/node", nil, false)
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
	setup(t, "https://test.com/node", clock, false)
	h1 := hasher.AddMock("fred")
	h1.ExpectTimestamp(clock.Times[0])
	h1.ExpectString("https://test.com/msg1\n")
	h1.ExpectString("hash")
	tx1 := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	h1.ExpectString(("https://user1.com/\n"))
	h1.ExpectSignature(tx1.Signatories[0].Signature)
	r.ResolveTx(tx1)
	tx2 := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
	h1.ExpectString(("https://user2.com/\n"))
	h1.ExpectSignature(tx2.Signatories[1].Signature)
	signer.Expect(types.Signature("tx-sig"), nodeKey, types.Hash("fred"))
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
	setup(t, "https://test.com/node", clock, false)
	h1 := hasher.AddMock("fred")
	h1.AcceptAnything()
	tx1 := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	r.ResolveTx(tx1)
	tx2 := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
	signer.SignAnythingAs("hello")
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

func TestTheReturnedTxIsSigned(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(t, "https://test.com/node", clock, false)
	tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/", true)
	stx, _ := records.CreateStoredTransaction(clock, &helpers.SHA512Factory{}, helpers.RSASigner{}, nodeKey, tx)
	if stx.Publisher == nil {
		t.Fatalf("the stored transaction was not signed")
	}
	err := rsa.VerifyPSS(&nodeKey.PublicKey, crypto.SHA512, stx.TxID, stx.Publisher.Signature, nil)
	if err != nil {
		t.Fatalf("signature verification failed")
	}
}

func TestSubmittingACompleteTransactionStoresItImmediately(t *testing.T) {
	clock := helpers.ClockDoubleIsoTimes("2024-12-25_03:00:00.121")
	setup(t, "https://test.com/node", clock, false)
	h1 := hasher.AddMock("fred")
	h1.AcceptAnything()
	signer.SignAnythingAs("hello")
	tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/", true)
	stx, _ := r.ResolveTx(tx)
	if stx == nil {
		t.Fatalf("the transaction was not stored")
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
	if !bytes.Equal(sigA.Signature, sigB.Signature) {
		t.Fatalf("Signature for %d did not match: %x not %x", which, sigA.Signature, sigB.Signature)
	}
}
