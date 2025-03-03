package storage

import (
	"crypto/rsa"
	"log"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type RemoteStorer interface {
	Handle(stx *records.StoredTransaction) error
}

type CheckAndStore struct {
	hasher  helpers.HasherFactory
	signer  helpers.Signer
	key     *rsa.PublicKey
	journal Journaller
}

func (cas *CheckAndStore) Handle(stx *records.StoredTransaction) error {
	log.Printf("asked to check and store remote tx\n")
	err := stx.VerifySignature(cas.hasher, cas.signer, cas.key)
	if err != nil {
		return err
	}
	return nil
}

func NewRemoteStorer(hasher helpers.HasherFactory, signer helpers.Signer, key *rsa.PublicKey, journal Journaller) RemoteStorer {
	return &CheckAndStore{hasher: hasher, signer: signer, key: key, journal: journal}
}
