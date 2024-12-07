package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

func main() {
	log.Println("starting chainledger")
	storeRecord := clienthandler.NewRecordStorage()
	cliapi := http.NewServeMux()
	cliapi.Handle("/store", storeRecord)
	err := http.ListenAndServe(":5001", cliapi)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}
