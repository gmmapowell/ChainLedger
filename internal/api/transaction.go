package api

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/url"
	"slices"
	"sync/atomic"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Transaction struct {
	ContentLink *url.URL
	ContentHash types.Hash
	Signatories []*types.Signatory
	completed   atomic.Bool
}

func NewTransaction(linkStr string, h types.Hash) (*Transaction, error) {
	link, err := url.Parse(linkStr)
	if err != nil {
		return nil, err
	}

	return &Transaction{ContentLink: link, ContentHash: h, Signatories: make([]*types.Signatory, 0)}, nil
}

func (tx *Transaction) ID() types.Hash {
	hasher := sha512.New()
	hasher.Write([]byte(tx.ContentLink.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(tx.ContentHash)
	for _, s := range tx.Signatories {
		hasher.Write([]byte(s.Signer.String()))
		hasher.Write([]byte("\n"))
	}
	return hasher.Sum(nil)
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

	for i, s := range tx.Signatories {
		if s.Signer.String() > signer.Signer.String() {
			tx.Signatories = slices.Insert(tx.Signatories, i, signer)
			return nil
		} else if s.Signer.String() == signer.Signer.String() {
			return fmt.Errorf("duplicate signer: %s", signer.Signer.String())
		}
	}

	tx.Signatories = append(tx.Signatories, signer)
	return nil
}

func (tx *Transaction) Sign(signerURL *url.URL, pk *rsa.PrivateKey) error {
	if signerURL == nil {
		return fmt.Errorf("no signer id provided")
	}
	if pk == nil {
		return fmt.Errorf("no private key provided")
	}
	return tx.doSign(signerURL, pk, nil)
}

func (tx *Transaction) doSign(signer *url.URL, pk *rsa.PrivateKey, e1 error) error {
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
	var h = sha512.New()
	h.Write([]byte("hello, world"))
	return h, nil
}

func makeSignature(pk *rsa.PrivateKey, h hash.Hash) (types.Signature, error) {
	sum := h.Sum(nil)
	sig, err := rsa.SignPSS(rand.Reader, pk, crypto.SHA512, sum, nil)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (tx *Transaction) JsonReader() (io.Reader, error) {
	json, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(json), nil
}

func (tx *Transaction) String() string {
	sigs := ""
	for _, s := range tx.Signatories {
		sigs = sigs + ", " + s.Signer.String()
	}
	return fmt.Sprintf("Tx[%s, %v%s]", tx.ContentLink, tx.ContentHash, sigs)
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	var m = make(map[string]any)
	m["ContentLink"] = tx.ContentLink.String()
	m["ContentHash"] = tx.ContentHash
	m["Signatories"] = tx.Signatories
	return json.Marshal(m)
}

func (tx *Transaction) UnmarshalJSON(bs []byte) error {
	var wire struct {
		ContentLink string
		ContentHash []byte
		Signatories []*types.Signatory
	}
	if err := json.Unmarshal(bs, &wire); err != nil {
		return err
	}
	if url, err := url.Parse(wire.ContentLink); err == nil {
		tx.ContentLink = url
	} else {
		return err
	}
	tx.ContentHash = wire.ContentHash
	tx.Signatories = wire.Signatories

	return nil
}

func (tx *Transaction) AlreadyCompleted() bool {
	return tx.completed.Swap(true)
}
