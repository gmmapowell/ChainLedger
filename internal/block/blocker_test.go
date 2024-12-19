package block_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/block"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

func TestBuildingBlock0(t *testing.T) {
	nodeName, _ := url.Parse("https://node1.com")
	pk, _ := rsa.GenerateKey(rand.Reader, 32)
	hasher := helpers.NewMockHasherFactory(t)
	buildTo, _ := types.ParseTimestamp("2024-12-12_18:00:00.000")
	mock1 := hasher.AddMock("computed-hash")
	mock1.ExpectString(nodeName.String() + "\n")
	mock1.ExpectTimestamp(buildTo)
	signer := helpers.MockSigner{}
	retHash := types.Hash([]byte("computed-hash"))
	retSig := types.Hash([]byte("signed as"))
	signer.Expect(retSig, pk, retHash)
	blocker := block.NewBlocker(hasher, &signer, nodeName, pk)
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
