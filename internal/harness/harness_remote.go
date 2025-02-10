package harness

import (
	"crypto/rsa"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/storage"
)

type HarnessRemote struct {
	from   *HarnessNode
	public *rsa.PublicKey
}

// Name implements config.NodeConfig.
func (h *HarnessRemote) Name() *url.URL {
	url, err := url.Parse(h.from.Name)
	if err != nil {
		panic("could not parse url " + h.from.Name)
	}
	return url
}

// PublicKey implements config.NodeConfig.
func (h *HarnessRemote) PublicKey() *rsa.PublicKey {
	return h.public
}

func (h *HarnessRemote) Handler() storage.RemoteStorer {
	return nil
}
