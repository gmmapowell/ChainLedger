package records

import (
	"hash"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type StoredTransaction struct {
	txid         hash.Hash
	whenReceived types.Timestamp
	contentLink  url.URL
	contentHash  hash.Hash
	signatories  []types.Signatory
}
