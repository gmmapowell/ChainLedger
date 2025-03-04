package block

import (
	"crypto/rsa"
	"log"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Blocker struct {
	hasher helpers.HasherFactory
	signer helpers.Signer
	name   *url.URL
	pk     *rsa.PrivateKey
}

func (b Blocker) Build(to types.Timestamp, last *records.Block, txs []*records.StoredTransaction) (*records.Block, error) {
	ls := "<none>"
	var lastID types.Hash
	if last != nil {
		ls = last.String()
		lastID = last.ID
	}
	log.Printf("Building block before %s, following %s with %d records\n", to.IsoTime(), ls, len(txs))

	txids := make([]types.Hash, len(txs))
	for i, tx := range txs {
		txids[i] = tx.TxID
	}

	block := &records.Block{
		UpUntil: to,
		BuiltBy: b.name,
		PrevID:  lastID,
		Txs:     txids,
	}

	var err error
	block.ID = block.HashMe(b.hasher)
	block.Signature, err = b.signer.Sign(b.pk, types.Hash(block.ID))
	if err != nil {
		return nil, err
	}

	return block, nil
}

func NewBlocker(hasher helpers.HasherFactory, signer helpers.Signer, name *url.URL, pk *rsa.PrivateKey) *Blocker {
	return &Blocker{hasher: hasher, signer: signer, name: name, pk: pk}
}
