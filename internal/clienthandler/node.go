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
	"github.com/gmmapowell/ChainLedger/internal/internode"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Node interface {
	Start() error
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

func (node *ListenerNode) Start() error {
	log.Printf("starting chainledger node %s\n", node.Name())
	nodeUrl, err := url.Parse(node.Name())
	if err != nil {
		log.Printf("could not parse node name as url: %s\n", node.Name())
		return err
	}
	clock := &helpers.ClockLive{}
	hasher := &helpers.SHA512Factory{}
	signer := &helpers.RSASigner{Name: nodeUrl}
	pending := storage.NewMemoryPendingStorage()
	resolver := NewResolver(clock, hasher, signer, node.config.PrivateKey(), pending)
	node.journaller = storage.NewJournaller(node.Name())
	node.runBlockBuilder(clock, node.journaller, node.config)
	node.startAPIListener(resolver, node.journaller)
	return nil
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
	senders := make([]internode.BinarySender, len(node.config.OtherNodes()))
	for i, n := range node.config.OtherNodes() {
		senders[i] = internode.NewHttpBinarySender(n.Name())
	}
	storeRecord := NewRecordStorage(resolver, journaller, senders)
	cliapi.Handle("/store", storeRecord)
	remoteTxHandler := internode.NewTransactionHandler(node.config)
	cliapi.Handle("/remotetx", remoteTxHandler)
	node.server = &http.Server{Addr: node.config.ListenOn(), Handler: cliapi}
	err := node.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(config config.LaunchableNodeConfig) Node {
	return &ListenerNode{config: config, Control: make(types.PingBack)}
}
