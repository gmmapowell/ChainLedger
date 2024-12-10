package storage

import "github.com/gmmapowell/ChainLedger/internal/api"

type PendingStorage interface {
	PendingTx(*api.Transaction) *api.Transaction
}

type MemoryPendingStorage struct {
	store map[int]*api.Transaction
}

func (mps MemoryPendingStorage) PendingTx(tx *api.Transaction) *api.Transaction {
	curr := mps.store[0]
	if curr == nil {
		mps.store[0] = tx
	}
	return curr
}

func NewMemoryPendingStorage() PendingStorage {
	return &MemoryPendingStorage{store: make(map[int]*api.Transaction)}
}
