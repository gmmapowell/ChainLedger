package records

import (
	"crypto/sha512"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type StoredTransaction struct {
	txid         sha512.Hash
	whenReceived types.Timestamp
	contentLink  url.URL
	contentHash  sha512.Hash
	signatories  []types.Signatory
}
