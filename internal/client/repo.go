package client

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/url"
)

type ClientRepository interface {
	PrivateKey(user *url.URL) (*rsa.PrivateKey, error)
}

type ClientInfo struct {
	user       *url.URL
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type MemoryClientRepository struct {
	clients map[url.URL]*ClientInfo
}

func MakeMemoryRepo() (ClientRepository, error) {
	mcr := MemoryClientRepository{clients: make(map[url.URL]*ClientInfo)}
	mcr.NewUser("https://user1.com/")
	mcr.NewUser("https://user2.com/")
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
	cr.clients[*u] = &ClientInfo{user: u, privateKey: pk, publicKey: &pk.PublicKey}
	return nil
}
