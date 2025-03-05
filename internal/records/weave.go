package records

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/url"

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

func UnmarshalBinaryWeave(bytes []byte) (*Weave, *types.Signatory, error) {
	weave := Weave{}
	buf := types.NewBinaryUnmarshallingBuffer(bytes)
	var err error
	weave.ID, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, nil, err
	}
	weave.ConsistentAt, err = types.UnmarshalTimestampFrom(buf)
	if err != nil {
		return nil, nil, err
	}
	weave.PrevID, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, nil, err
	}
	nblks, err := types.UnmarshalInt32From(buf)
	if err != nil {
		return nil, nil, err
	}
	weave.LatestBlocks = make([]NodeBlock, nblks)
	for i := 0; i < int(nblks); i++ {
		weave.LatestBlocks[i], err = UnmarshalBinaryNodeBlock(buf)
		if err != nil {
			return nil, nil, err
		}
	}

	signer := types.Signatory{}
	cls, err := types.UnmarshalStringFrom(buf)
	if err != nil {
		return nil, nil, err
	}
	signer.Signer, err = url.Parse(cls)
	if err != nil {
		return nil, nil, err
	}
	signer.Signature, err = types.UnmarshalSignatureFrom(buf)
	if err != nil {
		return nil, nil, err
	}

	err = buf.ShouldBeDone()
	if err != nil {
		return nil, nil, err
	}

	return &weave, &signer, nil
}

func (w *Weave) VerifySignatureIs(hasher helpers.HasherFactory, signer helpers.Signer, pub *rsa.PublicKey, signature types.Signature) error {
	id := w.HashMe(hasher)
	if !id.Is(w.ID) {
		return fmt.Errorf("remote weave id %s was not the result of computing it locally: %s", w.ID.String(), id.String())
	}
	return signer.Verify(pub, id, signature)
}
