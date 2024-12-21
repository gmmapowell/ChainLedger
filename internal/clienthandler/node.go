package clienthandler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/block"
	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type Node interface {
	Start()
}

type ListenerNode struct {
	name *url.URL
	addr string
}

func (node *ListenerNode) Start() {
	log.Println("starting chainledger node")
	clock := &helpers.ClockLive{}
	hasher := &helpers.SHA512Factory{}
	signer := &helpers.RSASigner{}
	config, err := config.ReadNodeConfig(node.name, node.addr)
	if err != nil {
		fmt.Printf("error reading config: %s\n", err)
		return
	}
	pending := storage.NewMemoryPendingStorage()
	resolver := NewResolver(clock, hasher, signer, config.NodeKey, pending)
	journaller := storage.NewJournaller()
	node.runBlockBuilder(clock, journaller, config)
	node.startAPIListener(resolver, journaller)
}

func (node ListenerNode) runBlockBuilder(clock helpers.Clock, journaller storage.Journaller, config *config.NodeConfig) {
	builder := block.NewBlockBuilder(clock, journaller, config.Name, config.NodeKey)
	builder.Start()
}

func (node *ListenerNode) startAPIListener(resolver Resolver, journaller storage.Journaller) {
	cliapi := http.NewServeMux()
	pingMe := PingHandler{}
	cliapi.Handle("/ping", pingMe)
	storeRecord := NewRecordStorage(resolver, journaller)
	cliapi.Handle("/store", storeRecord)
	err := http.ListenAndServe(node.addr, cliapi)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(name *url.URL, addr string) Node {
	return &ListenerNode{name, addr}
}
