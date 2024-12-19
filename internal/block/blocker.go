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

func (b Blocker) Build(to types.Timestamp, last *records.Block, txs []records.StoredTransaction) (*records.Block, error) {
	ls := "<none>"
	if last != nil {
		ls = last.String()
	}
	log.Printf("Building block before %s, following %s with %d records\n", to.IsoTime(), ls, len(txs))

	hasher := b.hasher.NewHasher()
	hasher.Write([]byte(b.name.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(to.AsBytes())
	hash := hasher.Sum(nil)

	sig, err := b.signer.Sign(b.pk, (*types.Hash)(&hash))
	if err != nil {
		return nil, err
	}

	return &records.Block{
		ID:        hash,
		UpUntil:   to,
		BuiltBy:   b.name,
		PrevID:    nil,
		Txs:       nil,
		Signature: *sig,
	}, nil
}

func NewBlocker(hasher helpers.HasherFactory, signer helpers.Signer, name *url.URL, pk *rsa.PrivateKey) *Blocker {
	return &Blocker{hasher: hasher, signer: signer, name: name, pk: pk}
}
