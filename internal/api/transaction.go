package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"hash/maphash"
	"io"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Transaction struct {
	ContentLink *url.URL
	ContentHash hash.Hash
	Signatories []*types.Signatory
}

func NewTransaction(linkStr string, h hash.Hash) (*Transaction, error) {
	link, err := url.Parse(linkStr)
	if err != nil {
		return nil, err
	}

	return &Transaction{ContentLink: link, ContentHash: h, Signatories: make([]*types.Signatory, 0)}, nil
}

func (tx *Transaction) SignerId(signerId string) error {
	signer, err := types.OtherSignerId(signerId)
	return tx.addSigner(signer, err)
}

func (tx *Transaction) Signer(signerURL *url.URL) error {
	signer, err := types.OtherSignerURL(signerURL)
	return tx.addSigner(signer, err)
}

func (tx *Transaction) addSigner(signer *types.Signatory, err error) error {
	if err != nil {
		return err
	}
	tx.Signatories = append(tx.Signatories, signer)
	return nil
}

func (tx *Transaction) Sign(signerURL *url.URL, pk string) error {
	return tx.doSign(signerURL, pk, nil)
}

func (tx *Transaction) doSign(signer *url.URL, pk string, e1 error) error {
	if e1 != nil {
		return e1
	}
	h, e2 := tx.makeSignableHash()
	if e2 != nil {
		return e2
	}
	sign, e3 := makeSignature(pk, h)
	if e3 != nil {
		return e3
	}
	done := false
	for _, signatory := range tx.Signatories {
		if signatory.Signer == signer {
			signatory.Signature = sign
			done = true
			break
		}
	}
	if !done {
		return fmt.Errorf("there is no signatory %v", signer)
	}
	return nil
}

func (tx *Transaction) makeSignableHash() (hash.Hash, error) {
	var h maphash.Hash
	h.WriteString("hello, world")
	return &h, nil
}

func makeSignature(pk string, h hash.Hash) (*types.Signature, error) {
	var r types.Signature = []byte(pk)
	r = h.Sum(r)
	return &r, nil
}

func (tx *Transaction) JsonReader() (io.Reader, error) {
	json, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(json), nil
}
