package storage

import (
	"log"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type JournalCommand interface {
}

type JournalStoreCommand struct {
	Tx *records.StoredTransaction
}

type JournalBlockCommand struct {
	Block *records.Block
}

type JournalRetrieveCommand struct {
	From, Upto types.Timestamp
	ResultChan chan<- []*records.StoredTransaction
}

type JournalHasBlockCommand struct {
	ID         types.Hash
	ResultChan chan<- bool
}

type JournalCheckTxsCommand struct {
	IDs        []types.Hash
	ResultChan chan<- []types.Hash
}

type JournalHasWeaveAtCommand struct {
	When       types.Timestamp
	ResultChan chan<- bool
}

type JournalStoreWeaveCommand struct {
	Weave *records.Weave
}

type JournalLatestBlockByCommand struct {
	Before     types.Timestamp
	ResultChan chan<- types.Hash
}

type JournalCheckCapacityCommand struct {
	AtLeast    int
	ResultChan chan<- bool
}

type JournalDoneCommand struct {
	NotifyMe chan<- struct{}
}

func LaunchJournalThread(name string, onNode string, finj helpers.FaultInjection) chan<- JournalCommand {
	var txs []*records.StoredTransaction
	var blocks []*records.Block
	weaves := make(map[types.Timestamp]*records.Weave)
	ret := make(chan JournalCommand, 20)
	log.Printf("%s: launching new journal thread for %s\n", onNode, name)
	go func() {
	whenDone:
		for {
			x := <-ret
			switch v := x.(type) {
			case JournalStoreCommand:
				txs = append(txs, v.Tx)
				log.Printf("%s recording tx with id %v, have %d at %p\n", name, v.Tx.TxID, len(txs), txs)
			case JournalBlockCommand:
				blocks = append(blocks, v.Block)
				log.Printf("%s recording block with id %v, have %d at %p\n", name, v.Block.ID, len(blocks), blocks)
			case JournalRetrieveCommand:
				log.Printf("reading txs = %p, len = %d\n", txs, len(txs))
				var ret []*records.StoredTransaction
				for _, tx := range txs {
					if tx.WhenReceived >= v.From && tx.WhenReceived < v.Upto {
						ret = append(ret, tx)
					}
				}
				v.ResultChan <- ret
			case JournalCheckCapacityCommand:
				ret := cap(txs) == len(txs) && cap(txs) >= v.AtLeast
				log.Printf("checking capacity, returning %v\n", ret)
				v.ResultChan <- ret
			case JournalHasBlockCommand:
				for _, b := range blocks {
					if b.ID.Is(v.ID) {
						v.ResultChan <- true
						continue whenDone
					}
				}
				v.ResultChan <- false
			case JournalCheckTxsCommand:
				tmp := make([]types.Hash, len(v.IDs))
				copy(tmp, v.IDs)
			nextTx:
				// go through all the TXs we _do_ have
				for _, tx := range txs {
					// is it in the list they are asking about
					for pos, id := range tmp {
						if id.Is(tx.TxID) {
							// if so, remove it and move on
							tmp[pos] = tmp[len(tmp)-1]
							tmp = tmp[:len(tmp)-1]
							continue nextTx
						}
					}
				}
				if len(tmp) == 0 {
					v.ResultChan <- nil
				} else {
					v.ResultChan <- tmp
				}
			case JournalHasWeaveAtCommand:
				v.ResultChan <- weaves[v.When] != nil
			case JournalStoreWeaveCommand:
				weaves[v.Weave.ConsistentAt] = v.Weave
			case JournalLatestBlockByCommand:
				var last types.Hash
				var lastWhen types.Timestamp
				for _, b := range blocks {
					if b.UpUntil <= v.Before && (last == nil || b.UpUntil > lastWhen) {
						last = b.ID
						lastWhen = b.UpUntil
					}
				}
				v.ResultChan <- last
			case JournalDoneCommand:
				log.Printf("was a done command %v\n", v)
				v.NotifyMe <- struct{}{}
				break whenDone
			default:
				log.Printf("not a valid journal command %v\n", x)
			}
		}
	}()
	return ret
}
