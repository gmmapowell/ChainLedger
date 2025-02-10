package storage

import (
	"crypto/rsa"
	"log"

	"github.com/gmmapowell/ChainLedger/internal/records"
)

type RemoteStorer interface {
	Handle(stx *records.StoredTransaction) error
}

type CheckAndStore struct {
	key     *rsa.PublicKey
	journal Journaller
}

func (cas *CheckAndStore) Handle(stx *records.StoredTransaction) error {
	log.Printf("asked to check and store remote tx\n")
	return nil
}

func NewRemoteStorer(key *rsa.PublicKey, journal Journaller) RemoteStorer {
	return &CheckAndStore{key: key, journal: journal}
}
