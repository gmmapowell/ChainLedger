package storage

import "github.com/gmmapowell/ChainLedger/internal/api"

type PendingStorage interface {
	PendingTx(*api.Transaction) *api.Transaction
}

type MemoryPendingStorage struct {
}

func (mps MemoryPendingStorage) PendingTx(tx *api.Transaction) *api.Transaction {
	return tx
}

func NewMemoryPendingStorage() PendingStorage {
	return new(MemoryPendingStorage)
}
