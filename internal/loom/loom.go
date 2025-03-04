package loom

import (
	"log"

	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Loom struct {
	myname string
}

func (loom *Loom) WeaveAt(when types.Timestamp) *records.Weave {
	log.Printf("%s weaving at %v\n", loom.myname, when)
	ret := records.Weave{ConsistentAt: when}
	return &ret
}
