package harness

import (
	"fmt"
	"hash/maphash"
	"log"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

type Config interface {
}

type Client interface {
	Begin()
	WaitFor()
}

type ConfigClient struct {
	submitter *client.Submitter
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
	}()
}

func (cli *ConfigClient) WaitFor() {

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
		ret[0] = &ConfigClient{submitter: s}
	}
	return ret
}
