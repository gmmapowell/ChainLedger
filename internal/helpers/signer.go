package helpers

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Signer interface {
	Sign(pk *rsa.PrivateKey, hash types.Hash) (types.Signature, error)
	SignerName() *url.URL
	Verify(pub *rsa.PublicKey, hash types.Hash, sig types.Signature) error
}

type RSASigner struct {
	Name *url.URL
}

func (s RSASigner) SignerName() *url.URL {
	return s.Name
}

func (s *RSASigner) Sign(pk *rsa.PrivateKey, hash types.Hash) (types.Signature, error) {
	sig, err := rsa.SignPSS(rand.Reader, pk, crypto.SHA512, []byte(hash), nil)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (s *RSASigner) Verify(pub *rsa.PublicKey, hash types.Hash, sig types.Signature) error {
	return rsa.VerifyPSS(pub, crypto.SHA512, hash, sig, nil)
}

type MockSigner struct {
	t    Fatals
	name *url.URL
	sigs []*MockExpectedSig
	next int
}

func (s MockSigner) SignerName() *url.URL {
	return s.name
}

func NewMockSigner(t Fatals, name *url.URL) *MockSigner {
	return &MockSigner{t: t, name: name}
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
		panic("there are no signatures")
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

func (s *MockSigner) Verify(pub *rsa.PublicKey, hash types.Hash, sig types.Signature) error {
	panic("unimplemented")
}

type MockExpectedSig struct {
	t         Fatals
	pk        *rsa.PrivateKey
	hash      types.Hash
	signature types.Signature
	accept    bool
}
