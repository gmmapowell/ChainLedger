package harness

import (
	"log"
	"sync"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/client"
)

type Client interface {
	PingNode()
	Begin()
	WaitFor()
}

type ConfigClient struct {
	repo      client.ClientRepository
	submitter *client.Submitter
	user      string
	count     int
	cosigners map[string]chan<- PleaseSign
	signFor   []<-chan PleaseSign
	done      chan struct{}
}

func (cli ConfigClient) PingNode() {
	cnt := 0
	for {
		err := cli.submitter.Ping()
		if err == nil {
			return
		}
		log.Println("ping failed")
		if cnt >= 10 {
			panic(err)
		}
		cnt++
		time.Sleep(1 * time.Second)
	}
}

func (cli *ConfigClient) Begin() {
	// We need to coordinate activity across all the cosigner threads, so create a waitgroup
	var wg sync.WaitGroup

	// Create one goroutine for each of the other clients attached to "this" node which might ask us to cosign for them
	for _, c := range cli.signFor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ps := range c {
				tx, err := makeTransaction(ps, cli.user)
				if err != nil {
					log.Fatal(err)
					continue
				}
				err = cli.submitter.Submit(tx)
				if err != nil {
					log.Fatal(err)
					continue
				}
			}
		}()
	}

	go func() {
		// Publish all of "our" messages
		for i := 0; i < cli.count; i++ {
			ps, err := makeMessage(cli)
			if err != nil {
				log.Fatal(err)
				continue
			}
			tx, err := makeTransaction(ps, cli.user)
			if err != nil {
				log.Fatal(err)
				continue
			}
			err = cli.submitter.Submit(tx)
			if err != nil {
				log.Fatal(err)
				continue
			}
			for _, u := range ps.cosigners {
				cli.cosigners[u.String()] <- ps
			}
		}

		// Tell all of the remote cosigners that we have finished by closing the channels
		for _, c := range cli.cosigners {
			close(c)
		}

		// Make sure all of our cosigners have finished
		wg.Wait()

		// Now we can report that we are fully done
		cli.done <- struct{}{}
	}()
}

// Wait for the Begin goroutine to signal that it is fully done
func (cli *ConfigClient) WaitFor() {
	<-cli.done
}
