package block

import (
	"crypto/rsa"
	"log"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Blocker struct {
	name *url.URL
	pk   *rsa.PrivateKey
}

func (b Blocker) Build(to types.Timestamp, last *records.Block, txs []records.StoredTransaction) *records.Block {
	ls := "<none>"
	if last != nil {
		ls = last.String()
	}
	log.Printf("Building block before %s, following %s with %d records\n", to.IsoTime(), ls, len(txs))

	return &records.Block{}
}
