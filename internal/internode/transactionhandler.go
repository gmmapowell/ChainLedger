package internode

import (
	"io"
	"log"
	"net/http"
)

type TransactionHandler struct {
}

// ServeHTTP implements http.Handler.
func (t *TransactionHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	buf, _ := io.ReadAll(req.Body)
	log.Printf("have received an internode request length: %d\n", len(buf))
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}
