package clienthandler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type RecordStorage struct {
	resolver Resolver
	journal  storage.Journaller
}

func NewRecordStorage(r Resolver, j storage.Journaller) RecordStorage {
	return RecordStorage{resolver: r, journal: j}
}

// ServeHTTP implements http.Handler.
func (r RecordStorage) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	log.Printf("asked to store record with length %d\n", req.ContentLength)

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error: %v\n", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("have json input %s\n", string(body))

	var tx = api.Transaction{}
	err = json.Unmarshal(body, &tx)
	if err != nil {
		log.Printf("Error unmarshalling: %v\n", err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("Have transaction %v\n", &tx)
	if stx, err := r.resolver.ResolveTx(&tx); stx != nil {
		r.journal.RecordTx(stx)
	} else if err != nil {
		log.Printf("Error resolving tx: %v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("have acknowledged this transaction, but not yet ready")
	}
}
