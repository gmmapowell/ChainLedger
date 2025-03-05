package storage

import (
	"log"

	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type WeaveConsolidator struct {
	commandChan chan<- WeaveConsolidationCommand
	stableChan  chan<- bool
	onNode      string
	nodeCount   int
}

type WeaveConsolidationCommand interface{}

type WeaveCreatedLocally struct {
	when types.Timestamp
	id   types.Hash
}

type WeaveSigned struct {
	when      types.Timestamp
	id        types.Hash
	by        string
	signature types.Signature
}

type WeaveAndSignatures struct {
	when       types.Timestamp
	id         types.Hash
	signatures map[string]types.Signature
}

func (wc *WeaveConsolidator) consolidate(ch <-chan WeaveConsolidationCommand) {
	consolidation := make(map[types.Timestamp]*WeaveAndSignatures)
	for {
		cmd := <-ch
		switch v := cmd.(type) {
		case WeaveCreatedLocally:
			log.Printf("%s: consolidating weave for %d\n", wc.onNode, v.when)
			if consolidation[v.when] == nil {
				consolidation[v.when] = &WeaveAndSignatures{when: v.when, id: v.id, signatures: make(map[string]types.Signature)}
			} else {
				log.Printf("cannot create weave for %d more than once\n", v.when)
			}
		case WeaveSigned:
			log.Printf("%s: consolidating signature by %s for weave for %d\n", wc.onNode, v.by, v.when)
			if consolidation[v.when] != nil {
				addSig := consolidation[v.when]
				if addSig.signatures[v.by] != nil {
					log.Printf("cannot add signature to weave for %d by %s more than once\n", v.when, v.by)
				} else if !addSig.id.Is(v.id) {
					log.Printf("cannot add signature to weave for %d by %s because hash values do not match\n", v.when, v.by)
				} else {
					addSig.signatures[v.by] = v.signature
					if wc.stableChan != nil && len(addSig.signatures) == wc.nodeCount {
						wc.stableChan <- true
					}
				}
			} else {
				log.Printf("cannot sign weave for %d yet, because it has not been created locally\n", v.when)
			}
		default:
			log.Printf("there is no case for command %v", v)
		}
	}
}

func (wc *WeaveConsolidator) LocalWeave(w *records.Weave) {
	cmd := WeaveCreatedLocally{when: w.ConsistentAt, id: w.ID}
	wc.commandChan <- cmd
}

func (wc *WeaveConsolidator) SignedWeave(when types.Timestamp, id types.Hash, by string, sig types.Signature) {
	cmd := WeaveSigned{when: when, id: id, by: by, signature: sig}
	wc.commandChan <- cmd
}

func (wc *WeaveConsolidator) NotifyMeWhenStable(ch chan<- bool) {
	wc.stableChan = ch
}

func NewWeaveConsolidator(onNode string, nodeCount int) *WeaveConsolidator {
	log.Printf("%s: creating new consolidator for %d nodes", onNode, nodeCount)
	ch := make(chan WeaveConsolidationCommand, 20)
	wc := &WeaveConsolidator{onNode: onNode, commandChan: ch, nodeCount: nodeCount}
	go wc.consolidate(ch)
	return wc
}
