package main

import (
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

func main() {
	url, _ := url.Parse("https://localhost:5001")
	node := clienthandler.NewListenerNode(url, ":5001")
	node.Start()
}
