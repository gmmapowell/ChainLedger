package main

import (
	"fmt"
	"hash/maphash"
	"log"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
)

func main() {
	cli, err := client.NewSubmitter("https://user1.com/", "private-key")
	if err != nil {
		log.Fatal(err)
		return
	}
	var h maphash.Hash
	h.WriteString("hello, world")
	tx, err := api.NewTransaction("http://tx.info/msg1", &h)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = tx.Signer("https://user2.com")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = cli.Submit(tx)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("submitted transaction: %v", tx)
}
