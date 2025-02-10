package block_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/block"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

var nodeName *url.URL
var pk *rsa.PrivateKey
var hasher *helpers.MockHasherFactory
var buildTo types.Timestamp
var mock1 *helpers.MockHasher
var signer *helpers.MockSigner
var retHash, prevID types.Hash
var retSig types.Signature
var blocker *block.Blocker

func setup(t *testing.T) {
	nodeName, _ = url.Parse("https://node1.com")
	pk, _ = rsa.GenerateKey(rand.Reader, 32)
	hasher = helpers.NewMockHasherFactory(t)
	buildTo, _ = types.ParseTimestamp("2024-12-12_18:00:00.000")
	mock1 = hasher.AddMock("computed-hash")
	signer = helpers.NewMockSigner(t, nodeName)
	prevID = types.Hash([]byte("previous-block"))
	retHash = types.Hash([]byte("computed-hash"))
	retSig = types.Signature([]byte("signed as"))
	signer.Expect(retSig, pk, retHash)
	blocker = block.NewBlocker(hasher, signer, nodeName, pk)
}

func TestBuildingBlock0(t *testing.T) {
	setup(t)
	mock1.ExpectString(nodeName.String() + "\n")
	mock1.ExpectTimestamp(buildTo)
	block0, _ := blocker.Build(buildTo, nil, nil)
	if block0.PrevID != nil {
		t.Fatalf("Block0 should have a nil previous block")
	}
	if block0.UpUntil != buildTo {
		t.Fatalf("the stored block time was not correct")
	}
	if len(block0.Txs) != 0 {
		t.Fatalf("Block0 should not have any messages")
	}
	if !bytes.Equal(block0.ID, retHash) {
		t.Fatalf("the computed hash was incorrect")
	}
	if !bytes.Equal(block0.Signature, retSig) {
		t.Logf("expected sig: %v\n", retSig)
		t.Logf("actual sig:   %v\n", block0.Signature)
		t.Fatalf("the computed signature was incorrect")
	}
}

func TestBuildingSubsequentBlockWithNoMessages(t *testing.T) {
	setup(t)
	mock1.ExpectString(string(prevID))
	mock1.ExpectString(nodeName.String() + "\n")
	mock1.ExpectTimestamp(buildTo)
	prev := records.Block{ID: prevID}
	block0, _ := blocker.Build(buildTo, &prev, nil)
	if !bytes.Equal(block0.PrevID, prevID) {
		t.Fatalf("Block1 should have a previous block id %v, not %v", prevID, block0.PrevID)
	}
	if block0.UpUntil != buildTo {
		t.Fatalf("the stored block time was not correct")
	}
	if len(block0.Txs) != 0 {
		t.Fatalf("Block0 should not have any messages")
	}
	if !bytes.Equal(block0.ID, retHash) {
		t.Fatalf("the computed hash was incorrect")
	}
	if !bytes.Equal(block0.Signature, retSig) {
		t.Logf("expected sig: %v\n", retSig)
		t.Logf("actual sig:   %v\n", block0.Signature)
		t.Fatalf("the computed signature was incorrect")
	}
}

func TestBuildingSubsequentBlockWithTwoMessages(t *testing.T) {
	setup(t)
	mock1.ExpectString(string(prevID))
	mock1.ExpectString(nodeName.String() + "\n")
	mock1.ExpectTimestamp(buildTo)

	prev := records.Block{ID: prevID}
	m1id := types.Hash([]byte("msg1"))
	m2id := types.Hash([]byte("msg2"))
	mock1.ExpectHash(m1id)
	mock1.ExpectHash(m2id)

	msg1 := records.StoredTransaction{TxID: m1id}
	msg2 := records.StoredTransaction{TxID: m2id}
	block0, _ := blocker.Build(buildTo, &prev, []*records.StoredTransaction{&msg1, &msg2})
	if !bytes.Equal(block0.PrevID, prevID) {
		t.Fatalf("Block1 should have a previous block id %v, not %v", prevID, block0.PrevID)
	}
	if block0.UpUntil != buildTo {
		t.Fatalf("the stored block time was not correct")
	}
	if len(block0.Txs) != 0 {
		t.Fatalf("Block0 should not have any messages")
	}
	if !bytes.Equal(block0.ID, retHash) {
		t.Fatalf("the computed hash was incorrect")
	}
	if !bytes.Equal(block0.Signature, retSig) {
		t.Logf("expected sig: %v\n", retSig)
		t.Logf("actual sig:   %v\n", block0.Signature)
		t.Fatalf("the computed signature was incorrect")
	}
}
