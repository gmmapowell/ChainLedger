package harness

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/url"
	"os"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type Config interface {
	NodeNames() []string
	Launcher(forNode string) config.LaunchableNodeConfig
	Remote(forNode string) config.NodeConfig
	ClientsFor(forNode string) []*CliConfig
}

type HarnessConfig struct {
	WeaveInterval int
	Nodes         []*HarnessNode
	keys          map[string]*rsa.PrivateKey
	pubs          map[string]*rsa.PublicKey
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
			return &HarnessLauncher{config: c, launching: n, private: c.keys[n.Name], public: &c.keys[n.Name].PublicKey, handlers: makeRemoteHandlers(c, n.Name)}
		}
	}
	panic("no node found for " + forNode)
}

func makeRemoteHandlers(c *HarnessConfig, name string) map[string]storage.RemoteStorer {
	hf := helpers.SHA512Factory{}
	sf := helpers.RSASigner{}

	ret := make(map[string]storage.RemoteStorer)
	for _, remote := range c.Nodes {
		if remote.Name == name {
			continue
		}
		ret[remote.Name] = storage.NewRemoteStorer(hf, &sf, c.pubs[remote.Name], storage.NewJournaller(remote.Name))
	}
	return ret
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
	ret.pubs = make(map[string]*rsa.PublicKey)

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
		ret.pubs[name] = &pk.PublicKey
	}

	return &ret
}
