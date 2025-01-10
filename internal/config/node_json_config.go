package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"

	"io"
	"net/url"
	"os"
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
}

// ListenOn implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) ListenOn() string {
	return n.config.ListenOn
}

// Name implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) Name() *url.URL {
	return n.url
}

// OtherNodes implements LaunchableNodeConfig.
func (n *NodeConfigWrapper) OtherNodes() []NodeConfig {
	return n.others
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

		others[i] = &NodeConfigWrapper{config: json, public: pub}
	}
	return &NodeConfigWrapper{config: config, url: url, private: pk, public: &pk.PublicKey, others: others}
}
