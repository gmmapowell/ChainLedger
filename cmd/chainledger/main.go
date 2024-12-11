package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

func main() {
	log.Println("starting chainledger")
	config, err := config.ReadNodeConfig()
	if err != nil {
		fmt.Printf("error reading config: %s\n", err)
		return
	}
	pending := storage.NewMemoryPendingStorage()
	resolver := clienthandler.NewResolver(&helpers.ClockLive{}, config.NodeKey, pending)
	journaller := storage.NewJournaller()
	storeRecord := clienthandler.NewRecordStorage(resolver, journaller)
	cliapi := http.NewServeMux()
	cliapi.Handle("/store", storeRecord)
	err = http.ListenAndServe(":5001", cliapi)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}
