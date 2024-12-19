package helpers

import (
	"crypto/sha512"
	"hash"
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
	hashers []MockHasher
	next    int
}

func (f *MockHasherFactory) AddMock(hashesTo string) MockHasher {
	ret := MockHasher{hashesTo: hashesTo}
	f.hashers = append(f.hashers, ret)
	return ret
}

func (f *MockHasherFactory) NewHasher() hash.Hash {
	r := f.hashers[f.next]
	f.next++
	return r
}

type MockHasher struct {
	hashesTo string
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
	return []byte(m.hashesTo)
}

// Write implements hash.Hash.
func (m MockHasher) Write(p []byte) (n int, err error) {
	panic("unimplemented")
}
