package storage

import (
	"crypto/rsa"
	"fmt"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type RemoteStorer interface {
	StoreTx(stx *records.StoredTransaction) error
	StoreBlock(block *records.Block) error
}

type CheckAndStore struct {
	hasher  helpers.HasherFactory
	signer  helpers.Signer
	key     *rsa.PublicKey
	journal Journaller
}

func (cas *CheckAndStore) StoreTx(stx *records.StoredTransaction) error {
	err := stx.VerifySignature(cas.hasher, cas.signer, cas.key)
	if err != nil {
		return err
	}
	return cas.journal.RecordTx(stx)
}

func (cas *CheckAndStore) StoreBlock(block *records.Block) error {
	err := block.VerifySignature(cas.hasher, cas.signer, cas.key)
	if err != nil {
		return err
	}
	hasBlock := cas.journal.HasBlock(block.PrevID)
	if !hasBlock {
		return fmt.Errorf("block %v does not have prev %v", block.ID, block.PrevID)
	}
	missingTxs := cas.journal.CheckTxs(block.Txs)
	if missingTxs != nil {
		return fmt.Errorf("block %v does not have %d txs", block.ID, len(missingTxs))
	}
	return cas.journal.RecordBlock(block)
}

func NewRemoteStorer(hasher helpers.HasherFactory, signer helpers.Signer, key *rsa.PublicKey, journal Journaller) RemoteStorer {
	return &CheckAndStore{hasher: hasher, signer: signer, key: key, journal: journal}
}
