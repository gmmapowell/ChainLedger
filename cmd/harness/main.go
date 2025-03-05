package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/harness"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: harness <json>\n")
		return
	}
	log.Println("starting harness")

	config := harness.ReadConfig(os.Args[1])
	nodes := harness.StartNodes(config)
	clients := harness.PrepareClients(config)

	startedAt := time.Now().UnixMilli()

	for _, c := range clients {
		c.Begin()
	}

	for _, c := range clients {
		c.WaitFor()
	}

	for _, n := range nodes {
		n.ClientsDone()
	}

	handsUp := make([]chan bool, len(nodes))
	for k, n := range config.NodeNames() {
		launcher := config.Launcher(n)
		handsUp[k] = make(chan bool)
		launcher.Consolidator().NotifyMeWhenStable(handsUp[k])
	}
	timeout := time.After(5 * time.Second)
outer:
	for k, c := range handsUp {
		select {
		case <-timeout:
			log.Printf("did not consolidate after 5s")
			break outer
		case worked := <-c:
			log.Printf("consolidator %d notified me: %v", k, worked)
		}
	}

	for _, n := range nodes {
		n.Terminate()
	}

	endedAt := time.Now().UnixMilli()
	log.Printf("elapsed time = %d", endedAt-startedAt)

	log.Println("harness complete")
}
