package records

import (
	"hash"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type NodeBlock struct {
	NodeName      string
	LatestBlockID types.Hash
}

func (n *NodeBlock) HashInto(hasher hash.Hash) {
	hasher.Write([]byte(n.NodeName))
	hasher.Write([]byte("\n"))
	hasher.Write(n.LatestBlockID)
}
