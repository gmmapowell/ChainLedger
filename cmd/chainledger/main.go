package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: chainledger <config>")
		return
	}
	config := config.ReadNodeConfig(os.Args[1])
	node := clienthandler.NewListenerNode(config)
	err := node.Start()
	if err != nil {
		log.Fatalf("node did not start: %v", err)
	}
}
