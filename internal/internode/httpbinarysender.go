package internode

import (
	"bytes"
	"log"
	"net/http"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/helpers"
)

type HttpBinarySender struct {
	cli *http.Client
	url *url.URL
}

// Send implements BinarySender.
func (h *HttpBinarySender) Send(path string, blob []byte) {
	tourl := h.url.JoinPath(path).String()
	log.Printf("sending blob(%d) to %s\n", len(blob), tourl)
	resp, err := h.cli.Post(tourl, "application/octet-stream", bytes.NewReader(blob))
	if err != nil {
		log.Printf("error sending to %s: %v\n", tourl, err)
	} else if resp.StatusCode/100 != 2 {
		log.Printf("bad status code sending to %s: %d\n", tourl, resp.StatusCode)
	}
}

func NewHttpBinarySender(url *url.URL) helpers.BinarySender {
	return &HttpBinarySender{cli: &http.Client{}, url: url}
}
