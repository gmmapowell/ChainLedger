package clienthandler_test

import (
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

var repo client.ClientRepository
var s storage.PendingStorage
var r clienthandler.Resolver

func setup() {
	repo, _ = client.MakeMemoryRepo()
	s = storage.NewMemoryPendingStorage()
	r = clienthandler.NewResolver(s)
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
	setup()
	tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
	stx, err := r.ResolveTx(tx)
	if stx != nil {
		t.Fatalf("a stored transaction was returned when the message was not fully signed")
	}
	if err != nil {
		t.Fatalf("ResolveTx returned an error: %v\n", err)
	}
}

func TestTwoCopiesOfTheTransactionAreEnoughToContinue(t *testing.T) {
	setup()
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", true, "https://user2.com/")
		r.ResolveTx(tx)
	}
	var stx *records.StoredTransaction
	var err error
	{
		tx := maketx("https://test.com/msg1", "hash", "https://user1.com/", "https://user2.com/", true)
		r.ResolveTx(tx)
		stx, err = r.ResolveTx(tx)
	}
	if stx == nil {
		t.Fatalf("a stored transaction was not returned after both parties had submitted a signed copy")
	}
	if err != nil {
		t.Fatalf("ResolveTx returned an error: %v\n", err)
	}
}
