package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func main() {
	log.Println("starting chainledger")
	pending := storage.NewMemoryPendingStorage()
	resolver := clienthandler.NewResolver(&helpers.ClockLive{}, pending)
	storeRecord := clienthandler.NewRecordStorage(resolver)
	cliapi := http.NewServeMux()
	cliapi.Handle("/store", storeRecord)
	err := http.ListenAndServe(":5001", cliapi)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}
