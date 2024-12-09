package clienthandler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/api"
)

type RecordStorage struct {
	resolver Resolver
}

func NewRecordStorage(r Resolver) RecordStorage {
	return RecordStorage{resolver: r}
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

	log.Printf("Have transaction %v\n", tx)
	if stx, err := r.resolver.ResolveTx(&tx); stx != nil {
		// TODO: move the transaction on to the next stage
		log.Printf("TODO: move it next stage")
	} else if err != nil {
		log.Printf("Error resolving tx: %v\n", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		log.Printf("have acknowledged this transaction, but not yet ready")
	}
}
