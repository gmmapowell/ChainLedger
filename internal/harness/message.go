package harness

import (
	"crypto/sha512"
	rno "math/rand/v2"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type PleaseSign struct {
	content    string
	hash       types.Hash
	originator url.URL
	cosigners  []url.URL
}

// Create a random message and return it as a PleaseSign
func makeMessage(cli *ConfigClient) (PleaseSign, error) {
	content := "http://tx.info/" + randomPath()

	hasher := sha512.New()
	hasher.Write(randomBytes(16))
	h := hasher.Sum(nil)

	return PleaseSign{
		content:    content,
		hash:       h,
		originator: cli.repo.URLFor(cli.user),
		cosigners:  cli.repo.OtherThan(cli.user),
	}, nil
}

// Create a transaction from a PleaseSign request
func makeTransaction(ps PleaseSign, submitter string) (*api.Transaction, error) {
	tx, err := api.NewTransaction(ps.content, ps.hash)
	if err != nil {
		return nil, err
	}
	for _, s := range ps.cosigners {
		if s.String() == submitter {
			continue
		}
		err = tx.Signer(&s)
		if err != nil {
			return nil, err
		}
	}
	if ps.originator.String() != submitter {
		err = tx.Signer(&ps.originator)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// Generate a random string to use as the "unique" message path
func randomPath() string {
	ns := 6 + rno.IntN(6)
	ret := make([]rune, ns)
	for i := 0; i < ns; i++ {
		ret[i] = alnumRune()
	}
	return string(ret)
}

// Generate a random character from a-z._
func alnumRune() rune {
	r := rno.IntN(38)
	switch {
	case r == 0:
		return '_'
	case r == 1:
		return '.'
	case r >= 2 && r < 12:
		return rune('0' + r - 2)
	case r >= 12:
		return rune('a' + r - 12)
	}
	panic("this should be in the range 0-38")
}

// Generate a random set of bytes to be used as a hash
func randomBytes(ns int) []byte {
	ret := make([]byte, ns)
	for i := 0; i < ns; i++ {
		ret[i] = byte(rno.IntN(256))
	}
	return ret
}
