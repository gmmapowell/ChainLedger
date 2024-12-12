package harness

import (
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

type Config interface {
}

type Client interface {
	Begin()
	WaitFor()
}

func ReadConfig() *Config {
	return nil
}

func StartNodes(c *Config) {
	node := clienthandler.NewListenerNode(":5001")
	go node.Start()
}

func PrepareClients(c *Config) []Client {
	return make([]Client, 0)
}
