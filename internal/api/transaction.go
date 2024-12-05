package api

import (
	"hash"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Transaction struct {
	ContentLink *url.URL
	ContentHash hash.Hash
	Signatories []types.Signatory
}

func NewTransaction(linkStr string, h hash.Hash) (*Transaction, error) {
	var link, err = url.Parse(linkStr)
	if err != nil {
		return nil, err
	}

	return &Transaction{ContentLink: link, ContentHash: h}, nil
}

func (tx Transaction) Signer(signerId string) error {
	return nil
}
