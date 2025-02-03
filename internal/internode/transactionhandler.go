package internode

import (
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/records"
)

type TransactionHandler struct {
}

// ServeHTTP implements http.Handler.
func (t *TransactionHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("could not read the buffer from the request")
		return
	}
	log.Printf("have received an internode request length: %d\n", len(buf))
	stx, err := records.UnmarshalBinaryStoredTransaction(buf)
	if err != nil {
		log.Printf("could not unpack the internode message")
		return
	}
	log.Printf("unmarshalled message to: %v\n", stx)
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}
