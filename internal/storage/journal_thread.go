package storage

import "fmt"

type JournalCommand interface {
}

type JournalStoreCommand struct {
}

type JournalRetrieveCommand struct {
}

type JournalDoneCommand struct {
	NotifyMe chan<- struct{}
}

func LaunchJournalThread() chan<- JournalCommand {
	ret := make(chan JournalCommand)
	go func() {
	whenDone:
		for {
			x := <-ret
			switch v := x.(type) {
			case JournalStoreCommand:
				fmt.Printf("was a store command %v\n", v)
			case JournalRetrieveCommand:
				fmt.Printf("was a retrieve command %v\n", v)
			case JournalDoneCommand:
				fmt.Printf("was a done command %v\n", v)
				v.NotifyMe <- struct{}{}
				break whenDone
			default:
				fmt.Printf("not a valid journal command %v\n", x)
			}
		}
	}()
	return ret
}
