package harness

import (
	"crypto/rsa"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/config"
)

type HarnessLauncher struct {
	config    *HarnessConfig
	launching *HarnessNode
	private   *rsa.PrivateKey
	public    *rsa.PublicKey
}

// Name implements config.LaunchableNodeConfig.
func (h *HarnessLauncher) Name() *url.URL {
	return h.launching.url
}

// PublicKey implements config.LaunchableNodeConfig.
func (h *HarnessLauncher) PublicKey() *rsa.PublicKey {
	return &h.config.keys[h.launching.Name].PublicKey
}

// ListenOn implements config.LaunchableNodeConfig.
func (h *HarnessLauncher) ListenOn() string {
	return h.launching.ListenOn
}

// OtherNodes implements config.LaunchableNodeConfig.
func (h *HarnessLauncher) OtherNodes() []config.NodeConfig {
	ret := make([]config.NodeConfig, len(h.config.NodeNames())-1)
	j := 0
	for _, n := range h.config.NodeNames() {
		if n == h.launching.Name {
			continue
		}
		ret[j] = h.config.Remote(n)
		j++
	}
	return ret
}

// PrivateKey implements config.LaunchableNodeConfig.
func (h *HarnessLauncher) PrivateKey() *rsa.PrivateKey {
	return h.private
}