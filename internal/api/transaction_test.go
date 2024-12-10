package api_test

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

func TestTwoSignaturesAddedInCollatingOrderStayThatWay(t *testing.T) {
	tx, _ := api.NewTransaction("https://test.com", []byte("hashcode"))
	u1, _ := url.Parse("http://user1.com")
	tx.Signer(u1)
	u2, _ := url.Parse("http://user2.com")
	tx.Signer(u2)
	if tx.Signatories[0].Signer.Host != "user1.com" {
		t.Fatalf("the first signer was %s\n", tx.Signatories[0].Signer)
	}
	if tx.Signatories[1].Signer.Host != "user2.com" {
		t.Fatalf("the second signer was %s\n", tx.Signatories[1].Signer)
	}
}

func TestTwoSignaturesAddedInInverseCollatingOrderAreReversed(t *testing.T) {
	tx, _ := api.NewTransaction("https://test.com", []byte("hashcode"))
	u2, _ := url.Parse("http://user2.com")
	tx.Signer(u2)
	u1, _ := url.Parse("http://user1.com")
	tx.Signer(u1)
	if tx.Signatories[0].Signer.Host != "user1.com" {
		t.Fatalf("the first signer was %s\n", tx.Signatories[0].Signer)
	}
	if tx.Signatories[1].Signer.Host != "user2.com" {
		t.Fatalf("the second signer was %s\n", tx.Signatories[1].Signer)
	}
}

func TestTheSameSignatoryCannotBeAddedTwice(t *testing.T) {
	tx, _ := api.NewTransaction("https://test.com", []byte("hashcode"))
	u1, _ := url.Parse("http://user1.com")
	tx.Signer(u1)
	err := tx.Signer(u1)
	if err == nil {
		t.Fatalf("we were allowed to add the same signer twice")
	}
}

func TestTransactionsHaveDistinctIDs(t *testing.T) {
	all := make([]types.Hash, 0)
	options := [2]struct {
		l  string
		h  string
		u1 string
		u2 string
	}{
		{"https://test.com/tx1", "hash1", "https://user1.com", "https://user2.com"},
		{"https://test.com/tx2", "hash2", "https://user3.com", "https://user4.com"},
	}

	for i := 0; i < 16; i++ {
		tx, _ := api.NewTransaction(options[bit(3, i)].l, types.Hash([]byte(options[bit(2, i)].h)))
		u1, _ := url.Parse(options[bit(1, i)].u1)
		u2, _ := url.Parse(options[bit(0, i)].u2)
		tx.Signer(u1)
		tx.Signer(u2)
		fmt.Printf("idx %d: %v\n", i, tx)
		all = append(all, tx.ID())
	}

	for i := 0; i < 16; i++ {
		for j := i + 1; j < 16; j++ {
			if bytes.Equal(all[i], all[j]) {
				t.Fatalf("two transactions had the same ID: %d and %d", i, j)
			}
		}
	}
}

func bit(b int, v int) int {
	return (v >> b) & 0x1
}
