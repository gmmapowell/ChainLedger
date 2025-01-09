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
	hasher  helpers.HasherFactory
	signer  helpers.Signer
	nodeKey *rsa.PrivateKey
	store   storage.PendingStorage
	finj    helpers.FaultInjection
}

func (r TxResolver) ResolveTx(tx *api.Transaction) (*records.StoredTransaction, error) {
	curr := r.store.PendingTx(tx)
	r.finj.NextWaiter("resolve-tx")
	complete := true
	for i, v := range tx.Signatories {
		if v.Signature != nil && curr != nil {
			curr.Signatories[i].Signature = v.Signature
		} else if v.Signature == nil {
			if curr == nil || curr.Signatories[i].Signature == nil {
				complete = false
			}
		}
	}

	if complete {
		if curr == nil {
			curr = tx
		}
		if !curr.AlreadyCompleted() {
			return records.CreateStoredTransaction(r.clock, r.hasher, r.signer, r.nodeKey, curr)
		}
	}

	return nil, nil
}

func NewResolver(clock helpers.Clock, hasher helpers.HasherFactory, signer helpers.Signer, nodeKey *rsa.PrivateKey, store storage.PendingStorage) Resolver {
	return TestResolver(helpers.IgnoreFaultInjection(), clock, hasher, signer, nodeKey, store)
}

func TestResolver(finj helpers.FaultInjection, clock helpers.Clock, hasher helpers.HasherFactory, signer helpers.Signer, nodeKey *rsa.PrivateKey, store storage.PendingStorage) Resolver {
	return &TxResolver{finj: finj, clock: clock, hasher: hasher, signer: signer, nodeKey: nodeKey, store: store}
}
