package records

import (
	"net/url"

	"encoding/base64"

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
