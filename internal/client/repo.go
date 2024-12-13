package client

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	rno "math/rand/v2"
	"net/url"
)

type ClientRepository interface {
	PrivateKey(user *url.URL) (*rsa.PrivateKey, error)

	SubmitterFor(nodeId string, userId string) (*Submitter, error)

	OtherThan(userId string) []url.URL
}

type ClientInfo struct {
	user       *url.URL
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type MemoryClientRepository struct {
	clients map[url.URL]*ClientInfo
	users   []url.URL
}

func MakeMemoryRepo() (MemoryClientRepository, error) {
	mcr := MemoryClientRepository{clients: make(map[url.URL]*ClientInfo)}
	return mcr, nil
}

func (cr MemoryClientRepository) PrivateKey(user *url.URL) (pk *rsa.PrivateKey, e error) {
	entry := cr.clients[*user]
	if entry == nil {
		e = fmt.Errorf("there is no user %s", user.String())
	} else {
		pk = entry.privateKey
	}
	return
}

func (cr MemoryClientRepository) HasUser(user string) bool {
	u, _ := url.Parse(user)
	return u != nil && cr.clients[*u] != nil
}

func (cr *MemoryClientRepository) NewUser(user string) error {
	u, e1 := url.Parse(user)
	if e1 != nil {
		return e1
	}
	if cr.clients[*u] != nil {
		return fmt.Errorf("user %s already exists in the repo", user)
	}
	pk, e2 := rsa.GenerateKey(rand.Reader, 2048)
	if e2 != nil {
		return e2
	}
	cr.users = append(cr.users, *u)
	cr.clients[*u] = &ClientInfo{user: u, privateKey: pk, publicKey: &pk.PublicKey}
	return nil
}

func (cr MemoryClientRepository) SubmitterFor(nodeId string, userId string) (*Submitter, error) {
	if uu, err := url.Parse(userId); err != nil {
		return nil, err
	} else if pk, err := cr.PrivateKey(uu); err != nil {
		return nil, err
	} else {
		return NewSubmitter(nodeId, userId, pk)
	}
}

func (cr MemoryClientRepository) OtherThan(userId string) []url.URL {
	for {
		q := rno.IntN(len(cr.users))
		who := cr.users[q]
		if who.String() != userId {
			return []url.URL{who}
		}
	}
}
