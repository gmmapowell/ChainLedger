package harness

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/url"
	"os"

	"github.com/gmmapowell/ChainLedger/internal/config"
)

type Config interface {
	NodeNames() []string
	Launcher(forNode string) config.LaunchableNodeConfig
	Remote(forNode string) config.NodeConfig
	ClientsFor(forNode string) []*CliConfig
}

type HarnessConfig struct {
	Nodes []*HarnessNode
	keys  map[string]*rsa.PrivateKey
}

type HarnessNode struct {
	Name     string
	ListenOn string
	Clients  []*CliConfig
	url      *url.URL
}

type CliConfig struct {
	User  string
	Count int
}

// NodeEndpoints implements Config.
func (c *HarnessConfig) NodeNames() []string {
	ret := make([]string, len(c.Nodes))
	for i, n := range c.Nodes {
		ret[i] = n.Name
	}
	return ret
}

func (c *HarnessConfig) Launcher(forNode string) config.LaunchableNodeConfig {
	for _, n := range c.Nodes {
		if n.Name == forNode {
			return &HarnessLauncher{config: c, launching: n, private: c.keys[n.Name], public: &c.keys[n.Name].PublicKey}
		}
	}
	panic("no node found for " + forNode)
}

// Remote implements Config.
func (c *HarnessConfig) Remote(forNode string) config.NodeConfig {
	for _, n := range c.Nodes {
		if n.Name == forNode {
			return &HarnessRemote{from: n, public: &c.keys[forNode].PublicKey}
		}
	}
	panic("no node found for " + forNode)
}

// ClientsPerNode implements Config.
func (c *HarnessConfig) ClientsFor(forNode string) []*CliConfig {
	for _, n := range c.Nodes {
		if n.Name == forNode {
			return n.Clients
		}
	}
	panic("no node found for " + forNode)
}

func ReadConfig(file string) Config {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	bytes, _ := io.ReadAll(fd)
	var ret HarnessConfig
	json.Unmarshal(bytes, &ret)
	ret.keys = make(map[string]*rsa.PrivateKey)

	for _, n := range ret.Nodes {
		name := n.Name
		url, err := url.Parse(name)
		if err != nil {
			panic("could not parse name " + name)
		}
		n.url = url
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic("key generation failed")
		}

		ret.keys[name] = pk
	}

	return &ret
}
