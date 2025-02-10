package config

import (
	"crypto/rsa"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type NodeConfig interface {
	Name() *url.URL
	PublicKey() *rsa.PublicKey
}

type LaunchableNodeConfig interface {
	NodeConfig
	ListenOn() string
	PrivateKey() *rsa.PrivateKey
	OtherNodes() []NodeConfig
	RemoteStorer(string) storage.RemoteStorer
}
