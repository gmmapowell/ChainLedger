package clienthandler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

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
	config     config.LaunchableNodeConfig
	Control    types.PingBack
	server     *http.Server
	journaller storage.Journaller
}

func (node *ListenerNode) Name() string {
	return node.config.Name().String()
}

func (node *ListenerNode) Start() {
	log.Printf("starting chainledger node %s\n", node.Name())
	clock := &helpers.ClockLive{}
	hasher := &helpers.SHA512Factory{}
	signer := &helpers.RSASigner{}
	pending := storage.NewMemoryPendingStorage()
	resolver := NewResolver(clock, hasher, signer, node.config.PrivateKey(), pending)
	node.journaller = storage.NewJournaller(node.Name())
	node.runBlockBuilder(clock, node.journaller, node.config)
	node.startAPIListener(resolver, node.journaller)
}

func (node *ListenerNode) Terminate() {
	node.server.Shutdown(context.Background())
	waitChan := make(types.Signal)
	node.Control <- waitChan.Sender()
	<-waitChan
	node.journaller.Quit()
	log.Printf("node %s finished\n", node.Name())
}

func (node ListenerNode) runBlockBuilder(clock helpers.Clock, journaller storage.Journaller, config config.LaunchableNodeConfig) {
	builder := block.NewBlockBuilder(clock, journaller, config.Name(), config.PrivateKey(), node.Control)
	builder.Start()
}

func (node *ListenerNode) startAPIListener(resolver Resolver, journaller storage.Journaller) {
	cliapi := http.NewServeMux()
	pingMe := PingHandler{}
	cliapi.Handle("/ping", pingMe)
	storeRecord := NewRecordStorage(resolver, journaller)
	cliapi.Handle("/store", storeRecord)
	node.server = &http.Server{Addr: node.config.ListenOn(), Handler: cliapi}
	err := node.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(config config.LaunchableNodeConfig) Node {
	return &ListenerNode{config: config, Control: make(types.PingBack)}
}
