package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"io"
	"net/url"
	"os"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type NodeJsonConfig struct {
	Name       string
	ListenOn   string
	PrivateKey string
	PublicKey  string
	OtherNodes []NodeJsonConfig
}

type NodeConfigWrapper struct {
	config  NodeJsonConfig
	url     *url.URL
	private *rsa.PrivateKey
	public  *rsa.PublicKey
	others  []NodeConfig
	handler storage.RemoteStorer // only for remote nodes
}

// ListenOn implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) ListenOn() string {
	return n.config.ListenOn
}

func (n *NodeConfigWrapper) WeaveInterval() int {
	panic("not implemented")
}

// Name implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) Name() *url.URL {
	return n.url
}

// OtherNodes implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) OtherNodes() []NodeConfig {
	return n.others
}

func (n *NodeConfigWrapper) RemoteStorer(name string) storage.RemoteStorer {
	for _, rn := range n.others {
		if rn.Name().String() == name {
			return rn.(*NodeConfigWrapper).handler
		}
	}
	return nil
}

// PrivateKey implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) PrivateKey() *rsa.PrivateKey {
	return n.private
}

// PublicKey implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) PublicKey() *rsa.PublicKey {
	return n.public
}

func ReadNodeConfig(file string) LaunchableNodeConfig {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	bytes, _ := io.ReadAll(fd)
	var config NodeJsonConfig
	json.Unmarshal(bytes, &config)

	url, err := url.Parse(config.Name)
	if err != nil {
		panic("cannot parse url " + config.Name)
	}

	pkbs, err := base64.StdEncoding.DecodeString(config.PrivateKey)
	if err != nil {
		panic("cannot parse base64 private key " + config.PrivateKey)
	}
	pk, err := x509.ParsePKCS1PrivateKey(pkbs)
	if err != nil {
		panic("cannot parse private key after conversion from " + config.PrivateKey)
	}

	hf := helpers.SHA512Factory{}
	sf := helpers.RSASigner{}

	others := make([]NodeConfig, len(config.OtherNodes))
	for i, json := range config.OtherNodes {
		bs, err := base64.StdEncoding.DecodeString(json.PublicKey)
		if err != nil {
			panic("cannot parse base64 public key " + json.PublicKey)
		}
		pub, err := x509.ParsePKCS1PublicKey(bs)
		if err != nil {
			panic("cannot parse public key after conversion from " + json.PublicKey)
		}

		others[i] = &NodeConfigWrapper{config: json, public: pub, handler: storage.NewRemoteStorer(hf, &sf, pub, storage.NewJournaller(json.Name))}
	}
	return &NodeConfigWrapper{config: config, url: url, private: pk, public: &pk.PublicKey, others: others}
}
