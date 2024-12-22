package clienthandler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/block"
	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Node interface {
	Start()
	Terminate()
}

type ListenerNode struct {
	name    *url.URL
	addr    string
	Control types.PingBack
	server  *http.Server
}

func (node *ListenerNode) Start() {
	log.Printf("starting chainledger node %s\n", node.name)
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
	journaller := storage.NewJournaller(node.name.String())
	node.runBlockBuilder(clock, journaller, config)
	node.startAPIListener(resolver, journaller)
}

func (node *ListenerNode) Terminate() {
	node.server.Shutdown(context.Background())
	waitChan := make(types.Signal)
	node.Control <- waitChan.Sender()
	<-waitChan
	log.Printf("node %s finished\n", node.name)
}

func (node ListenerNode) runBlockBuilder(clock helpers.Clock, journaller storage.Journaller, config *config.NodeConfig) {
	builder := block.NewBlockBuilder(clock, journaller, config.Name, config.NodeKey, node.Control)
	builder.Start()
}

func (node *ListenerNode) startAPIListener(resolver Resolver, journaller storage.Journaller) {
	cliapi := http.NewServeMux()
	pingMe := PingHandler{}
	cliapi.Handle("/ping", pingMe)
	storeRecord := NewRecordStorage(resolver, journaller)
	cliapi.Handle("/store", storeRecord)
	node.server = &http.Server{Addr: node.addr, Handler: cliapi}
	err := node.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(name *url.URL, addr string) Node {
	return &ListenerNode{name: name, addr: addr, Control: make(types.PingBack)}
}
