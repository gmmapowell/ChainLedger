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
}

type HarnessConfig struct {
	nodeEndpoints []string
}

// NodeEndpoints implements Config.
func (c *HarnessConfig) NodeEndpoints() []string {
	return c.nodeEndpoints
}

func ReadConfig() Config {
	return &HarnessConfig{nodeEndpoints: []string{":5001", ":5002"}}
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
	ret := make([]Client, 2)
	if s, err := repo.SubmitterFor("http://localhost:5001", "https://user1.com/"); err != nil {
		panic(err)
	} else {
		url, _ := url.Parse("https://user2.com/")
		ret[0] = &ConfigClient{submitter: s, other: url, done: make(chan struct{})}
	}
	if s, err := repo.SubmitterFor("http://localhost:5001", "https://user2.com/"); err != nil {
		panic(err)
	} else {
		url, _ := url.Parse("https://user1.com/")
		ret[1] = &ConfigClient{submitter: s, other: url, done: make(chan struct{})}
	}

	for _, s := range ret {
		s.PingNode()
	}
	return ret
}
