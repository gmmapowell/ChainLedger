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

type JournalCheckCapacityCommand struct {
	AtLeast    int
	ResultChan chan<- bool
}

type JournalDoneCommand struct {
	NotifyMe chan<- struct{}
}

func LaunchJournalThread(name string, finj helpers.FaultInjection) chan<- JournalCommand {
	var txs []*records.StoredTransaction
	var blocks []*records.Block
	ret := make(chan JournalCommand, 20)
	log.Printf("launching new journal thread with channel %p\n", ret)
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
