package storage

import (
	"sync"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
)

type PendingStorage interface {
	PendingTx(*api.Transaction) *api.Transaction
}

type MemoryPendingStorage struct {
	mu    sync.Mutex
	store map[string]*api.Transaction
	finj  helpers.FaultInjection
}

func (mps *MemoryPendingStorage) PendingTx(tx *api.Transaction) *api.Transaction {
	mps.mu.Lock()
	defer mps.mu.Unlock()
	curr := mps.store[string(tx.ID())]
	mps.finj.NextWaiter()
	if curr == nil {
		mps.store[string(tx.ID())] = tx
	}
	return curr
}

func NewMemoryPendingStorage() PendingStorage {
	return TestMemoryPendingStorage(helpers.IgnoreFaultInjection())
}

func TestMemoryPendingStorage(finj helpers.FaultInjection) PendingStorage {
	return &MemoryPendingStorage{store: make(map[string]*api.Transaction), finj: finj}
}
