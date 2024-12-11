package storage

import (
	"github.com/gmmapowell/ChainLedger/internal/api"
)

type PendingStorage interface {
	PendingTx(*api.Transaction) *api.Transaction
}

type MemoryPendingStorage struct {
	store map[string]*api.Transaction
}

func (mps MemoryPendingStorage) PendingTx(tx *api.Transaction) *api.Transaction {
	curr := mps.store[string(tx.ID())]
	if curr == nil {
		mps.store[string(tx.ID())] = tx
	}
	return curr
}

func NewMemoryPendingStorage() PendingStorage {
	return &MemoryPendingStorage{store: make(map[string]*api.Transaction)}
}
