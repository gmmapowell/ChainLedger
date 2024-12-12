package main

import (
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

func main() {
	node := clienthandler.NewListenerNode(":5001")
	node.Start()
}
