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

func (n *NodeBlock) MarshalBinaryInto(into *types.BinaryMarshallingBuffer) error {
	types.MarshalStringInto(into, n.NodeName)
	n.LatestBlockID.MarshalBinaryInto(into)
	return nil
}

func UnmarshalBinaryNodeBlock(buf *types.BinaryUnmarshallingBuffer) (NodeBlock, error) {
	ret := NodeBlock{}
	var err error
	ret.NodeName, err = types.UnmarshalStringFrom(buf)
	if err != nil {
		return ret, err
	}
	ret.LatestBlockID, err = types.UnmarshalHashFrom(buf)

	return ret, err
}
