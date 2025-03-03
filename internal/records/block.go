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

func (b Block) String() string {
	return "Block[" + base64.StdEncoding.EncodeToString(b.ID) + "]"
}
