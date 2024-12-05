package client

import (
	"github.com/gmmapowell/ChainLedger/internal/api"
)

type Submitter struct {
}

func NewSubmitter(id string, pk string) (Submitter, error) {
	return Submitter{}, nil
}

func (s *Submitter) Submit(tx *api.Transaction) error {
	return nil
}
