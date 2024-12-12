package main

import (
	"fmt"
	"hash/maphash"
	"log"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
)

func main() {
	repo, e1 := client.MakeMemoryRepo()
	if e1 != nil {
		panic(e1)
	}
	cli, err := repo.SubmitterFor("http://localhost:5001", "https://user1.com/")
	if err != nil {
		panic(err)
	}

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
	err = cli.Submit(tx)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Printf("submitted transaction: %v", tx)
}
