package internode

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type BlockHandler struct {
	nodeConfig config.LaunchableNodeConfig
}

// ServeHTTP implements http.Handler.
func (t *BlockHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("could not read the buffer from the request")
		return
	}
	log.Printf("have received an internode block length: %d\n", len(buf))
	block, err := records.UnmarshalBinaryBlock(buf)
	if err != nil {
		log.Printf("could not unpack the internode block: %v\n", err)
		return
	}
	log.Printf("unmarshalled block message to: %v\n", block)
	publishedBy := block.BuiltBy.String()
	storer := t.nodeConfig.RemoteStorer(publishedBy)
	if storer == nil {
		log.Printf("could not find a handler for remote node %s\n", publishedBy)
		return
	}
	err = storer.StoreBlock(block)
	if err != nil {
		panic(fmt.Sprintf("failed to store remote transaction: %v", err))
	}
}

func NewBlockHandler(c config.LaunchableNodeConfig) *BlockHandler {
	return &BlockHandler{nodeConfig: c}
}
