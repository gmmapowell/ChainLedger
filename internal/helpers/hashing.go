package helpers

import (
	"bytes"
	"crypto/sha512"
	"hash"
	"testing"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type HasherFactory interface {
	NewHasher() hash.Hash
}

type SHA512Factory struct {
}

func (f SHA512Factory) NewHasher() hash.Hash {
	return sha512.New()
}

type MockHasherFactory struct {
	t       *testing.T
	hashers []*MockHasher
	next    int
}

func (f *MockHasherFactory) AddMock(hashesTo string) *MockHasher {
	ret := &MockHasher{t: f.t, hashesTo: hashesTo, accepting: false}
	f.hashers = append(f.hashers, ret)
	return ret
}

func (f *MockHasherFactory) NewHasher() hash.Hash {
	if f.next >= len(f.hashers) {
		f.t.Fatalf("The mock hasher does not have %d hashers configured", f.next+1)
	}
	r := f.hashers[f.next]
	f.next++
	return r
}

func NewMockHasherFactory(t *testing.T) *MockHasherFactory {
	return &MockHasherFactory{t: t}
}

type MockHasher struct {
	t         *testing.T
	hashesTo  string
	accepting bool
	blobs     []byte
	written   []byte
}

// BlockSize implements hash.Hash.
func (m MockHasher) BlockSize() int {
	panic("unimplemented")
}

// Reset implements hash.Hash.
func (m MockHasher) Reset() {
	panic("unimplemented")
}

// Size implements hash.Hash.
func (m MockHasher) Size() int {
	panic("unimplemented")
}

// Sum implements hash.Hash.
func (m MockHasher) Sum(b []byte) []byte {
	// This is an implementation detail
	// and can easily be checked by adding another argument to the constructor
	if b != nil {
		panic("mock always expects final block to be nil")
	}
	if !m.accepting && !bytes.Equal(m.blobs, m.written) {
		m.t.Log("the written blobs were not the expected blobs")
		m.t.Logf("expected: %v\n", m.blobs)
		m.t.Logf("written:  %v\n", m.written)
		m.t.Fail()
	}
	return []byte(m.hashesTo)
}

// Write implements hash.Hash.
func (m *MockHasher) Write(p []byte) (n int, err error) {
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *MockHasher) AcceptAnything() *MockHasher {
	m.accepting = true
	return m
}

func (m *MockHasher) ExpectString(s string) *MockHasher {
	m.blobs = append(m.blobs, []byte(s)...)
	return m
}

func (m *MockHasher) ExpectHash(h types.Hash) *MockHasher {
	m.blobs = append(m.blobs, []byte(h)...)
	return m
}

func (m *MockHasher) ExpectSignature(s types.Signature) *MockHasher {
	m.blobs = append(m.blobs, []byte(s)...)
	return m
}

func (m *MockHasher) ExpectTimestamp(ts types.Timestamp) *MockHasher {
	m.blobs = append(m.blobs, ts.AsBytes()...)
	return m
}
