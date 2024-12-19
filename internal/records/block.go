package records

import (
	"net/url"

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

func (b Block) String() string {
	return "Block"
}
