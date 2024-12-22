package config

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/url"
)

type NodeConfig struct {
	Name     *url.URL
	ListenOn string
	NodeKey  *rsa.PrivateKey
}

func (nc *NodeConfig) UnmarshalJSON(bs []byte) error {
	var wire struct {
		Name     string
		ListenOn string
	}
	if err := json.Unmarshal(bs, &wire); err != nil {
		return err
	}
	if url, err := url.Parse(wire.Name); err == nil {
		nc.Name = url
	} else {
		return err
	}
	nc.ListenOn = wire.ListenOn

	return nil
}

func ReadNodeConfig(name *url.URL, addr string) (*NodeConfig, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &NodeConfig{Name: name, ListenOn: addr, NodeKey: pk}, nil
}
