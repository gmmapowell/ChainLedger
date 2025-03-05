package records

import (
	"log"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Weave struct {
	ID           types.Hash
	ConsistentAt types.Timestamp
	PrevID       types.Hash
	LatestBlocks []NodeBlock
}

func (w *Weave) HashMe(hf helpers.HasherFactory) types.Hash {
	hasher := hf.NewHasher()
	hasher.Write(w.ConsistentAt.AsBytes())
	hasher.Write(w.PrevID)
	for _, m := range w.LatestBlocks {
		m.HashInto(hasher)
	}
	return hasher.Sum(nil)

}

func (w *Weave) LogMe(node string) {
	log.Printf("%s weave.ID           = %v\n", node, w.ID)
	log.Printf("%s weave.ConsistentAt = %v\n", node, w.ConsistentAt)
	log.Printf("%s weave.PrevID       = %v\n", node, w.PrevID)
	for i, nb := range w.LatestBlocks {
		log.Printf("%s weave.Block[%d]    = %s => %v\n", node, i, nb.NodeName, nb.LatestBlockID)
	}
}
