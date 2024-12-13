package harness

import (
	"crypto/sha512"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

type Config interface {
	NodeEndpoints() []string
	ClientsPerNode() map[string][]CliConfig
}

type HarnessConfig struct {
	nodeEndpoints []string
	clients       map[string][]CliConfig
}

// NodeEndpoints implements Config.
func (c *HarnessConfig) NodeEndpoints() []string {
	return c.nodeEndpoints
}

// ClientsPerNode implements Config.
func (c *HarnessConfig) ClientsPerNode() map[string][]CliConfig {
	return c.clients
}

type CliConfig struct {
	client string
	other  string
}

func ReadConfig() Config {
	return &HarnessConfig{nodeEndpoints: []string{":5001", ":5002"}, clients: map[string][]CliConfig{
		"http://localhost:5001": {
			CliConfig{client: "https://user1.com/", other: "https://user2.com/"},
			CliConfig{client: "https://user2.com/", other: "https://user1.com/"},
		},
		"http://localhost:5002": {
			CliConfig{client: "https://user1.com/", other: "https://user2.com/"},
			CliConfig{client: "https://user2.com/", other: "https://user1.com/"},
		},
	}}
}

type Client interface {
	PingNode()
	Begin()
	WaitFor()
}

type ConfigClient struct {
	submitter *client.Submitter
	other     *url.URL
	done      chan struct{}
}

func (cli ConfigClient) PingNode() {
	cnt := 0
	for {
		err := cli.submitter.Ping()
		if err == nil {
			return
		}
		log.Println("ping failed")
		if cnt >= 10 {
			panic(err)
		}
		cnt++
		time.Sleep(1 * time.Second)
	}
}

func (cli *ConfigClient) Begin() {
	go func() {
		hasher := sha512.New()
		hasher.Write([]byte("hello, world"))
		h := hasher.Sum(nil)

		tx, err := api.NewTransaction("http://tx.info/msg1", h)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = tx.Signer(cli.other)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = cli.submitter.Submit(tx)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("submitted transaction: %v", tx)
		cli.done <- struct{}{}
	}()
}

func (cli *ConfigClient) WaitFor() {
	<-cli.done
}

func StartNodes(c Config) {
	for _, ep := range c.NodeEndpoints() {
		node := clienthandler.NewListenerNode(ep)
		go node.Start()
	}
}

func PrepareClients(c Config) []Client {
	repo, err := client.MakeMemoryRepo()
	if err != nil {
		panic(err)
	}
	ret := make([]Client, 0)
	m := c.ClientsPerNode()
	for node, clis := range m {
		for _, cli := range clis {
			if s, err := repo.SubmitterFor(node, cli.client); err != nil {
				panic(err)
			} else {
				url, _ := url.Parse(cli.other)
				ret = append(ret, &ConfigClient{submitter: s, other: url, done: make(chan struct{})})
			}
		}
	}

	for _, s := range ret {
		s.PingNode()
	}
	return ret
}
