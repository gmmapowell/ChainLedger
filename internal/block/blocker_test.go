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
	hasher := helpers.MockHasherFactory{}
	hasher.AddMock("computed-hash")
	blocker := block.NewBlocker(&hasher, nodeName, pk)
	buildTo, _ := types.ParseTimestamp("2024-12-12_18:00:00.000")
	retHash := types.Hash([]byte("computed-hash"))
	retSig := types.Hash([]byte("signed as"))
	block0 := blocker.Build(buildTo, nil, nil)
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
		t.Fatalf("the computed signature was incorrect")
	}
}
