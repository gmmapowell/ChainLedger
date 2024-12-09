package main

import (
	"fmt"
	"hash/maphash"
	"log"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
)

func main() {
	repo, e1 := client.MakeMemoryRepo()
	if e1 != nil {
		panic(e1)
	}
	uid := "https://user1.com/"
	uu, e2 := url.Parse(uid)
	if e2 != nil {
		panic(e2)
	}
	pk, e3 := repo.PrivateKey(uu)
	if e3 != nil {
		panic(e3)
	}
	cli, err := client.NewSubmitter("http://localhost:5001", uid, pk)
	if err != nil {
		log.Fatal(err)
		return
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
