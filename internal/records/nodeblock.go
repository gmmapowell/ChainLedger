package records

import "github.com/gmmapowell/ChainLedger/internal/types"

type NodeBlock struct {
	NodeName      string
	LatestBlockID types.Hash
}
