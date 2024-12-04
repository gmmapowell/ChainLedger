package main

import (
	"fmt"
    "log"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
)

func main() {
	var x, err = url.Parse("https://hello.com/")
    if err != nil {
        log.Fatal(err)
    }
	var tx = api.Transaction{ContentLink: *x}
	fmt.Println(tx)
}
