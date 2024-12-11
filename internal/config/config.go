package config

import (
	"crypto/rand"
	"crypto/rsa"
)

type NodeConfig struct {
	NodeKey *rsa.PrivateKey
}

func ReadNodeConfig() (*NodeConfig, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &NodeConfig{NodeKey: pk}, nil
}
