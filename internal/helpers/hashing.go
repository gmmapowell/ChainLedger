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

func (f *MockHasherFactory) NewHasher() hash.Hash {
	r := f.hashers[f.next]
	f.next++
	return r
}

type MockHasher struct {
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
	panic("unimplemented")
}

// Write implements hash.Hash.
func (m MockHasher) Write(p []byte) (n int, err error) {
	panic("unimplemented")
}

