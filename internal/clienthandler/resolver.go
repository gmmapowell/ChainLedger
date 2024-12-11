package clienthandler

import (
	"crypto/rsa"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type Resolver interface {
	ResolveTx(tx *api.Transaction) (*records.StoredTransaction, error)
}

type TxResolver struct {
	clock   helpers.Clock
	nodeKey *rsa.PrivateKey
	store   storage.PendingStorage
}

func (r TxResolver) ResolveTx(tx *api.Transaction) (*records.StoredTransaction, error) {
	curr := r.store.PendingTx(tx)
	complete := true
	for i, v := range tx.Signatories {
		if v.Signature != nil && curr != nil {
			curr.Signatories[i] = v
		} else if v.Signature == nil {
			if curr == nil || curr.Signatories[i].Signature == nil {
				complete = false
			}
		}
	}

	if complete {
		return records.CreateStoredTransaction(r.clock, r.nodeKey, curr)
	}

	return nil, nil
}

func NewResolver(clock helpers.Clock, nodeKey *rsa.PrivateKey, store storage.PendingStorage) Resolver {
	return &TxResolver{clock: clock, nodeKey: nodeKey, store: store}
}
