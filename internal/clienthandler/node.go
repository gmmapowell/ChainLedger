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
	"github.com/gmmapowell/ChainLedger/internal/loom"
	"github.com/gmmapowell/ChainLedger/internal/storage"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Node interface {
	Start() error
	ClientsDone()
	Terminate()
}

type ListenerNode struct {
	config         config.LaunchableNodeConfig
	BlockerControl types.PingBack
	LoomControl    chan string
	waitChan       types.Signal
	server         *http.Server
	journaller     storage.Journaller
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
	senders := make([]helpers.BinarySender, len(node.config.OtherNodes()))
	for i, n := range node.config.OtherNodes() {
		senders[i] = internode.NewHttpBinarySender(n.Name())
	}
	node.journaller = node.config.AllJournals()[node.Name()]
	node.runBlockBuilder(clock, node.journaller, node.config, hasher, signer, senders)
	node.runLoom(clock, hasher, signer, node.config.AllJournals(), senders)
	node.startAPIListener(clock, resolver, node.journaller, senders)
	return nil
}

func (node *ListenerNode) ClientsDone() {
	node.waitChan = make(types.Signal)
	node.BlockerControl <- node.waitChan.Sender()
}

func (node *ListenerNode) Terminate() {
	node.LoomControl <- "Quit"
	node.server.Shutdown(context.Background())
	<-node.waitChan
	node.journaller.Quit()
	log.Printf("node %s finished\n", node.Name())
}

func (node ListenerNode) runBlockBuilder(clock helpers.Clock, journaller storage.Journaller, config config.LaunchableNodeConfig, hasher helpers.HasherFactory, signer helpers.Signer, senders []helpers.BinarySender) {
	builder := block.NewBlockBuilder(clock, journaller, config.Name(), config.PrivateKey(), node.BlockerControl, hasher, signer, senders)
	builder.Start()
}

func (node ListenerNode) runLoom(clock helpers.Clock, hasher helpers.HasherFactory, signer helpers.Signer, allJournals map[string]storage.Journaller, senders []helpers.BinarySender) {
	theloom := loom.NewLoom(hasher, node.Name(), allJournals)
	l := loom.NewLoomThread(clock, node.config.Name().String(), node.LoomControl, node.config.WeaveInterval(), theloom, node.journaller, signer, node.config.PrivateKey(), senders)
	l.Start()
}

func (node *ListenerNode) startAPIListener(clock helpers.Clock, resolver Resolver, journaller storage.Journaller, senders []helpers.BinarySender) {
	cliapi := http.NewServeMux()
	pingMe := PingHandler{}
	cliapi.Handle("/ping", pingMe)
	storeRecord := NewRecordStorage(resolver, journaller, senders)
	cliapi.Handle("/store", storeRecord)
	remoteTxHandler := internode.NewTransactionHandler(node.config)
	cliapi.Handle("/remotetx", remoteTxHandler)
	remoteBlockHandler := internode.NewBlockHandler(node.config)
	cliapi.Handle("/remoteblock", remoteBlockHandler)
	remoteWeaveHandler := internode.NewWeaveHandler(node.config, clock)
	cliapi.Handle("/remoteweave", remoteWeaveHandler)
	node.server = &http.Server{Addr: node.config.ListenOn(), Handler: cliapi}
	err := node.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("error starting server: %s\n", err)
	}
}

func NewListenerNode(config config.LaunchableNodeConfig) Node {
	return &ListenerNode{config: config, BlockerControl: make(types.PingBack), LoomControl: make(chan string)}
}
