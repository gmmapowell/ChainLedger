package config

import (
	"crypto/rand"
	"crypto/rsa"
	"net/url"
)

type NodeConfig struct {
	Name     *url.URL
	ListenOn string
	NodeKey  *rsa.PrivateKey
}

func ReadNodeConfig(name *url.URL, addr string) (*NodeConfig, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &NodeConfig{Name: name, ListenOn: addr, NodeKey: pk}, nil
}
