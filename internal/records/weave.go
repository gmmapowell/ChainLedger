package records

import "github.com/gmmapowell/ChainLedger/internal/types"

type Weave struct {
	ID           types.Hash
	ConsistentAt types.Timestamp
	PrevID       types.Hash
	LatestBlocks []NodeBlock
}
