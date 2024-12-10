package clienthandler_test

import (
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

func TestANewTransactionMayBeStoredButReturnsNothing(t *testing.T) {
	repo, _ := client.MakeMemoryRepo()
	s := storage.NewMemoryPendingStorage()
	r := clienthandler.NewResolver(s)
	tx, _ := api.NewTransaction("https://test.com/msg1", types.Hash([]byte("hash")))
	u1, _ := url.Parse("https://user1.com/")
	pk, _ := repo.PrivateKey(u1)
	u2, _ := url.Parse("https://user2.com/")
	tx.Signer(u1)
	tx.Signer(u2)
	tx.Sign(u1, pk)
	stx, err := r.ResolveTx(tx)
	if stx != nil {
		t.Fatalf("a stored transaction was returned when the message was not fully signed")
	}
	if err != nil {
		t.Fatalf("ResolveTx returned an error: %v\n", err)
	}
}
