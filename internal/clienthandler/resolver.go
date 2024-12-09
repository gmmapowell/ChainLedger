package clienthandler

import (
	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type Resolver interface {
	ResolveTx(tx *api.Transaction) (*records.StoredTransaction, error)
}

type TxResolver struct {
}

func (r TxResolver) ResolveTx(tx *api.Transaction) (*records.StoredTransaction, error) {
	return nil, nil
}

func NewResolver(store storage.PendingStorage) Resolver {
	return new(TxResolver)
}
