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
	Sign(pk *rsa.PrivateKey, hash types.Hash) (types.Signature, error)
}

type RSASigner struct {
}

func (s RSASigner) Sign(pk *rsa.PrivateKey, hash types.Hash) (types.Signature, error) {
	sig, err := rsa.SignPSS(rand.Reader, pk, crypto.SHA512, []byte(hash), nil)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

type MockSigner struct {
	t    *testing.T
	sigs []*MockExpectedSig
	next int
}

func NewMockSigner(t *testing.T) *MockSigner {
	return &MockSigner{t: t}
}

func (f *MockSigner) Expect(signature types.Signature, pk *rsa.PrivateKey, hash types.Hash) {
	ret := &MockExpectedSig{t: f.t, signature: signature, pk: pk, hash: hash, accept: false}
	f.sigs = append(f.sigs, ret)
}

func (f *MockSigner) SignAnythingAs(sig string) {
	ret := &MockExpectedSig{t: f.t, signature: types.Signature([]byte(sig)), accept: true}
	f.sigs = append(f.sigs, ret)
}

func (f *MockSigner) Sign(pk *rsa.PrivateKey, hash types.Hash) (types.Signature, error) {
	if f.next >= len(f.sigs) {
		f.t.Fatalf("there are not %d signers configured", f.next+1)
	}
	r := f.sigs[f.next]
	if !r.accept {
		if pk != r.pk { // this is a pointer comparison, which is almost undoubtedly valid for tests
			f.t.Log("primary keys did not match")
			f.t.Fail()
		}
		if !bytes.Equal(r.hash, hash) {
			f.t.Log("hash was not correct")
			f.t.Fail()
		}
	}
	f.next++
	return types.Signature(r.signature), nil
}

type MockExpectedSig struct {
	t         *testing.T
	pk        *rsa.PrivateKey
	hash      types.Hash
	signature types.Signature
	accept    bool
}
