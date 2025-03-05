package internode

import (
	"io"
	"log"
	"net/http"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type WeaveHandler struct {
	nodeConfig config.LaunchableNodeConfig
}

// ServeHTTP implements http.Handler.
func (t *WeaveHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("could not read the buffer from the request")
		return
	}
	log.Printf("%s: received an internode block length: %d\n", t.nodeConfig.Name(), len(buf))
	weave, signer, err := records.UnmarshalBinaryWeave(buf)
	if err != nil {
		log.Printf("could not unpack the internode weave: %v\n", err)
		return
	}
	log.Printf("unmarshalled weave message to: %v\n", weave)
	storer := t.nodeConfig.RemoteStorer(signer.Signer.String())
	if storer == nil {
		log.Printf("could not find a handler for remote node %s\n", signer.Signer.String())
		return
	}

	// Now we need to compare and record this
}

func NewWeaveHandler(c config.LaunchableNodeConfig) *WeaveHandler {
	return &WeaveHandler{nodeConfig: c}
}
