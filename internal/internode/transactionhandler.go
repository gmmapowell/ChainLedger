package internode

import (
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type TransactionHandler struct {
	nodeConfig config.LaunchableNodeConfig
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
		log.Printf("could not unpack the internode message: %v\n", err)
		return
	}
	log.Printf("unmarshalled message to: %v\n", stx)
	publishedBy := stx.Publisher.Signer.String()
	storer := t.nodeConfig.RemoteStorer(publishedBy)
	if storer == nil {
		log.Printf("could not find a handler for remote node %s\n", publishedBy)
		return
	}
	storer.Handle(stx)
}

func NewTransactionHandler(c config.LaunchableNodeConfig) *TransactionHandler {
	return &TransactionHandler{nodeConfig: c}
}
