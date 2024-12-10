package api_test

import (
	"net/url"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/api"
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
