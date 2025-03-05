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

func (w *Weave) MarshalAndSend(senders []helpers.BinarySender, node string, sig types.Signature) {
	blob, err := w.MarshalBinary(node, sig)
	if err != nil {
		log.Printf("Error marshalling weave: %v %v\n", w.ID, err)
		return
	}
	for _, bs := range senders {
		go bs.Send("/remoteweave", blob)
	}
}

func (w *Weave) MarshalBinary(node string, sig types.Signature) ([]byte, error) {
	ret := types.NewBinaryMarshallingBuffer()

	// Marshal in the things that "belong to" the weave
	w.ID.MarshalBinaryInto(ret)
	w.ConsistentAt.MarshalBinaryInto(ret)
	w.PrevID.MarshalBinaryInto(ret)
	types.MarshalInt32Into(ret, int32(len(w.LatestBlocks)))
	for _, nb := range w.LatestBlocks {
		nb.MarshalBinaryInto(ret)
	}

	// and now marshal in the name and signature
	types.MarshalStringInto(ret, node)
	sig.MarshalBinaryInto(ret)

	return ret.Bytes(), nil
}
