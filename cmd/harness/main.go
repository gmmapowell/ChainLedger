package main

import (
	"log"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/harness"
)

func main() {
	log.Println("starting harness")

	config := harness.ReadConfig()
	harness.StartNodes(config)
	clients := harness.PrepareClients(config)

	startedAt := time.Now().UnixMilli()

	for _, c := range clients {
		c.Begin()
	}

	for _, c := range clients {
		c.WaitFor()
	}

	endedAt := time.Now().UnixMilli()
	log.Printf("elapsed time = %d", endedAt-startedAt)

	log.Println("harness complete")
}
