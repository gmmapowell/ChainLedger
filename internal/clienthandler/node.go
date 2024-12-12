package clienthandler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type Node interface {
	Start()
}

type ListenerNode struct {
	addr string
}

func (node *ListenerNode) Start() {
	log.Println("starting chainledger node")
	config, err := config.ReadNodeConfig()
	if err != nil {
		fmt.Printf("error reading config: %s\n", err)
		return
	}
	pending := storage.NewMemoryPendingStorage()
	resolver := NewResolver(&helpers.ClockLive{}, config.NodeKey, pending)
	journaller := storage.NewJournaller()
	storeRecord := NewRecordStorage(resolver, journaller)
	cliapi := http.NewServeMux()
	cliapi.Handle("/store", storeRecord)
	err = http.ListenAndServe(node.addr, cliapi)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(addr string) Node {
	return &ListenerNode{addr}
}
