package harness

import (
	"fmt"
	"hash/maphash"
	"log"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

type Config interface {
}

type Client interface {
	PingNode()
	Begin()
	WaitFor()
}

type ConfigClient struct {
	submitter *client.Submitter
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
		var hasher maphash.Hash
		hasher.WriteString("hello, world")
		h := hasher.Sum(nil)

		tx, err := api.NewTransaction("http://tx.info/msg1", h)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = tx.SignerId("https://user2.com")
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

func ReadConfig() *Config {
	return nil
}

func StartNodes(c *Config) {
	node := clienthandler.NewListenerNode(":5001")
	go node.Start()
}

func PrepareClients(c *Config) []Client {
	repo, err := client.MakeMemoryRepo()
	if err != nil {
		panic(err)
	}
	ret := make([]Client, 1)
	if s, err := repo.SubmitterFor("http://localhost:5001", "https://user1.com/"); err != nil {
		panic(err)
	} else {
		ret[0] = &ConfigClient{submitter: s, done: make(chan struct{})}

	for _, s := range ret {
		s.PingNode()
	}
	return ret
}
