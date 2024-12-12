package harness

import (
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
