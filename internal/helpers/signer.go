package helpers

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Signer interface {
	Sign(pk *rsa.PrivateKey, hash *types.Hash) (*types.Signature, error)
}

type RSASigner struct {
}

func (s RSASigner) Sign(pk *rsa.PrivateKey, hash *types.Hash) (*types.Signature, error) {
	sig, err := rsa.SignPSS(rand.Reader, pk, crypto.SHA512, []byte(*hash), nil)
	if err != nil {
		return nil, err
	}
	var ret types.Signature = sig
	return &ret, nil
}

type MockSigner struct {
	t    *testing.T
	sigs []*MockExpectedSig
	next int
}

func (f *MockSigner) Expect(signature types.Hash, pk *rsa.PrivateKey, hash types.Hash) {
	ret := &MockExpectedSig{t: f.t, signature: signature, pk: pk, hash: hash}
	f.sigs = append(f.sigs, ret)
}

func (f *MockSigner) Sign(pk *rsa.PrivateKey, hash *types.Hash) (*types.Signature, error) {
	r := f.sigs[f.next]
	if pk != r.pk { // this is a pointer comparison, which is almost undoubtedly valid for tests
		f.t.Log("primary keys did not match")
		f.t.Fail()
	}
	if !bytes.Equal(r.hash, *hash) {
		f.t.Log("hash was not correct")
		f.t.Fail()
	}
	f.next++
	return (*types.Signature)(&r.signature), nil
}

type MockExpectedSig struct {
	t         *testing.T
	pk        *rsa.PrivateKey
	hash      types.Hash
	signature []byte
}
