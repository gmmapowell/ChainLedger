package api

import (
	"hash"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Transaction struct {
	ContentLink url.URL
	ContentHash hash.Hash
	Signatories []types.Signatory
}
