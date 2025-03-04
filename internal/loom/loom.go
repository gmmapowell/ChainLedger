package loom

import (
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Loom struct {
}

func (loom *Loom) WeaveAt(when types.Timestamp) *records.Weave {
	ret := records.Weave{}
	return &ret
}
