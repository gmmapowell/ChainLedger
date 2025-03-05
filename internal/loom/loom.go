package loom

import (
	"sort"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Loom struct {
	myname      string
	allJournals map[string]storage.Journaller
	hf          helpers.HasherFactory
}

func (loom *Loom) Name() string {
	return loom.myname
}

func (loom *Loom) WeaveAt(when types.Timestamp, prev *records.Weave) *records.Weave {
	var prevID types.Hash
	if prev != nil {
		prevID = prev.ID
	}
	nbs := make([]records.NodeBlock, len(loom.allJournals))
	k := 0
	for n, j := range loom.allJournals {
		blk := j.LatestBlockBy(when)
		if len(blk) == 0 {
			return nil // it is not possible to weave if we don't have at least one block for every node
		}
		nbs[k] = records.NodeBlock{NodeName: n, LatestBlockID: blk}
		k++
	}
	sort.Slice(nbs, sortByName(nbs))
	ret := records.Weave{ConsistentAt: when, PrevID: prevID, LatestBlocks: nbs}
	ret.ID = ret.HashMe(loom.hf)
	return &ret
}

func sortByName(nbs []records.NodeBlock) func(i, j int) bool {
	return func(i, j int) bool {
		return nbs[i].NodeName < nbs[j].NodeName
	}
}

func NewLoom(hf helpers.HasherFactory, name string, allJournals map[string]storage.Journaller) *Loom {
	return &Loom{hf: hf, myname: name, allJournals: allJournals}
}
