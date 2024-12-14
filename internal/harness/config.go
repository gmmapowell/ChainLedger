package harness

import (
	"encoding/json"
	"io"
	"os"
)

type Config interface {
	NodeEndpoints() []string
	ClientsPerNode() map[string][]CliConfig
}

type HarnessConfig struct {
	Nodes   []string
	Clients map[string][]CliConfig
}

// NodeEndpoints implements Config.
func (c *HarnessConfig) NodeEndpoints() []string {
	return c.Nodes
}

// ClientsPerNode implements Config.
func (c *HarnessConfig) ClientsPerNode() map[string][]CliConfig {
	return c.Clients
}

type CliConfig struct {
	Client string `json:"user"`
	Count  int
}

func ReadConfig(file string) Config {
	fd, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer fd.Close()

	bytes, _ := io.ReadAll(fd)
	var ret HarnessConfig
	json.Unmarshal(bytes, &ret)

	return &ret
}
