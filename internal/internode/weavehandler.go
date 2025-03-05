package internode

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gmmapowell/ChainLedger/internal/config"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/records"
)

type WeaveHandler struct {
	nodeConfig config.LaunchableNodeConfig
	clock      helpers.Clock
}

const delay = 500 * time.Millisecond

// ServeHTTP implements http.Handler.
func (t *WeaveHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("could not read the buffer from the request")
		return
	}
	log.Printf("%s: received an internode block length: %d\n", t.nodeConfig.Name(), len(buf))
	weave, signatory, err := records.UnmarshalBinaryWeave(buf)
	if err != nil {
		log.Printf("could not unpack the internode weave: %v\n", err)
		return
	}
	log.Printf("unmarshalled weave message to: %v\n", weave)
	storer := t.nodeConfig.RemoteStorer(signatory.Signer.String())
	if storer == nil {
		log.Printf("could not find a handler for remote node %s\n", signatory.Signer.String())
		return
	}

	// Hack-ish: wait 500ms so that our local node has built its own
	timer := t.clock.After(delay)
	<-timer

	// Tell the storer for that node that we have this signature
	err = storer.SignedWeave(weave, signatory.Signature)
	if err != nil {
		panic(fmt.Sprintf("%s: cannot accept signed weave: %v", t.nodeConfig.Name(), err))
	}
}

func NewWeaveHandler(c config.LaunchableNodeConfig, clock helpers.Clock) *WeaveHandler {
	return &WeaveHandler{nodeConfig: c, clock: clock}
}
