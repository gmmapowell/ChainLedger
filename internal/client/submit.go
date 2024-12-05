package client

import (
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
)

type Submitter struct {
	iam *url.URL
	pk string
}

func NewSubmitter(id string, pk string) (*Submitter, error) {
	iam, err := url.Parse(id)
	if err != nil {
		return nil, err
	}
	return &Submitter{iam: iam, pk: pk}, nil
}

func (s *Submitter) Submit(tx *api.Transaction) error {
	tx.Signer(s.iam)
	tx.Sign(s.iam, s.pk)
	return nil
}
