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
