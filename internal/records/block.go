package records

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/url"

	"encoding/base64"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Block struct {
	ID        types.Hash
	PrevID    types.Hash
	BuiltBy   *url.URL
	UpUntil   types.Timestamp
	Txs       []types.Hash
	Signature types.Signature
}

func (b *Block) VerifySignature(hasher helpers.HasherFactory, signer helpers.Signer, pub *rsa.PublicKey) error {
	id := b.HashMe(hasher)
	if !id.Is(b.ID) {
		return fmt.Errorf("remote block id %s was not the result of computing it locally: %s", b.ID.String(), id.String())
	}
	return signer.Verify(pub, id, b.Signature)
}

func (b *Block) HashMe(hf helpers.HasherFactory) types.Hash {
	hasher := hf.NewHasher()
	hasher.Write(b.PrevID)
	hasher.Write([]byte(b.BuiltBy.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(b.UpUntil.AsBytes())
	for _, m := range b.Txs {
		hasher.Write(m)
	}
	return hasher.Sum(nil)
}

func (b *Block) MarshalAndSend(senders []helpers.BinarySender) {
	blob, err := b.MarshalBinary()
	if err != nil {
		log.Printf("Error marshalling block: %v %v\n", b.ID, err)
		return
	}
	for _, bs := range senders {
		go bs.Send("/remoteblock", blob)
	}
}

func (b *Block) MarshalBinary() ([]byte, error) {
	ret := types.NewBinaryMarshallingBuffer()
	b.ID.MarshalBinaryInto(ret)
	b.PrevID.MarshalBinaryInto(ret)
	types.MarshalStringInto(ret, b.BuiltBy.String())
	b.UpUntil.MarshalBinaryInto(ret)
	types.MarshalInt32Into(ret, int32(len(b.Txs)))
	for _, tx := range b.Txs {
		tx.MarshalBinaryInto(ret)
	}
	b.Signature.MarshalBinaryInto(ret)
	return ret.Bytes(), nil
}

func UnmarshalBinaryBlock(bytes []byte) (*Block, error) {
	block := Block{}
	buf := types.NewBinaryUnmarshallingBuffer(bytes)
	var err error
	block.ID, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, err
	}
	block.PrevID, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, err
	}
	cls, err := types.UnmarshalStringFrom(buf)
	if err != nil {
		return nil, err
	}
	block.BuiltBy, err = url.Parse(cls)
	if err != nil {
		return nil, err
	}
	block.UpUntil, err = types.UnmarshalTimestampFrom(buf)
	if err != nil {
		return nil, err
	}
	ntxs, err := types.UnmarshalInt32From(buf)
	if err != nil {
		return nil, err
	}
	block.Txs = make([]types.Hash, ntxs)
	for i := 0; i < int(ntxs); i++ {
		block.Txs[i], err = types.UnmarshalHashFrom(buf)
		if err != nil {
			return nil, err
		}
	}
	block.Signature, err = types.UnmarshalSignatureFrom(buf)
	if err != nil {
		return nil, err
	}
	err = buf.ShouldBeDone()
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (b Block) String() string {
	return "Block[" + base64.StdEncoding.EncodeToString(b.ID) + "]"
}
