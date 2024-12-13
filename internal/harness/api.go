package harness

import (
	"crypto/sha512"
	"log"
	"time"

	rno "math/rand/v2"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

type Config interface {
	NodeEndpoints() []string
	ClientsPerNode() map[string][]CliConfig
}

type HarnessConfig struct {
	nodeEndpoints []string
	clients       map[string][]CliConfig
}

// NodeEndpoints implements Config.
func (c *HarnessConfig) NodeEndpoints() []string {
	return c.nodeEndpoints
}

// ClientsPerNode implements Config.
func (c *HarnessConfig) ClientsPerNode() map[string][]CliConfig {
	return c.clients
}

type CliConfig struct {
	client string
	count  int
}

func ReadConfig() Config {
	return &HarnessConfig{nodeEndpoints: []string{":5001", ":5002"}, clients: map[string][]CliConfig{
		"http://localhost:5001": {
			CliConfig{client: "https://user1.com/", count: 10},
			CliConfig{client: "https://user2.com/", count: 2},
		},
		"http://localhost:5002": {
			CliConfig{client: "https://user1.com/", count: 5},
			CliConfig{client: "https://user2.com/", count: 7},
		},
	}}
}

type Client interface {
	PingNode()
	Begin()
	WaitFor()
}

type ConfigClient struct {
	repo      client.ClientRepository
	submitter *client.Submitter
	user      string
	count     int
	done      chan struct{}
}

func (cli ConfigClient) PingNode() {
	cnt := 0
	for {
		err := cli.submitter.Ping()
		if err == nil {
			return
		}
		log.Println("ping failed")
		if cnt >= 10 {
			panic(err)
		}
		cnt++
		time.Sleep(1 * time.Second)
	}
}

func (cli *ConfigClient) Begin() {
	go func() {
		for i := 0; i < cli.count; i++ {
			tx, err := makeMessage(cli)
			if err != nil {
				log.Fatal(err)
				return
			}
			err = cli.submitter.Submit(tx)
			if err != nil {
				log.Fatal(err)
				return
			}
		}
		cli.done <- struct{}{}
	}()
}

func makeMessage(cli *ConfigClient) (*api.Transaction, error) {
	content := "http://tx.info/" + randomPath()
	hasher := sha512.New()
	hasher.Write(randomBytes(16))
	h := hasher.Sum(nil)

	tx, err := api.NewTransaction(content, h)
	if err != nil {
		return nil, err
	}
	for _, s := range cli.repo.OtherThan(cli.user) {
		err = tx.Signer(&s)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (cli *ConfigClient) WaitFor() {
	<-cli.done
}

func StartNodes(c Config) {
	for _, ep := range c.NodeEndpoints() {
		node := clienthandler.NewListenerNode(ep)
		go node.Start()
	}
}

func PrepareClients(c Config) []Client {
	repo, err := client.MakeMemoryRepo()
	if err != nil {
		panic(err)
	}
	m := c.ClientsPerNode()
	for _, clis := range m {
		for _, cli := range clis {
			if repo.HasUser(cli.client) {
				continue
			} else if err := repo.NewUser(cli.client); err != nil {
				panic(err)
			}
		}
	}
	ret := make([]Client, 0)
	for node, clis := range m {
		for _, cli := range clis {
			if s, err := repo.SubmitterFor(node, cli.client); err != nil {
				panic(err)
			} else {
				ret = append(ret, &ConfigClient{repo: &repo, submitter: s, user: cli.client, count: cli.count, done: make(chan struct{})})
			}
		}
	}

	for _, s := range ret {
		s.PingNode()
	}
	return ret
}

func randomPath() string {
	ns := 6 + rno.IntN(6)
	ret := make([]rune, ns)
	for i := 0; i < ns; i++ {
		ret[i] = alnumRune()
	}
	return string(ret)
}

func alnumRune() rune {
	r := rno.IntN(38)
	switch {
	case r == 0:
		return '-'
	case r == 1:
		return '.'
	case r >= 2 && r < 12:
		return rune('0' + r - 2)
	case r >= 12:
		return rune('a' + r - 12)
	}
	panic("this should be in the range 0-38")
}

func randomBytes(ns int) []byte {
	ret := make([]byte, ns)
	for i := 0; i < ns; i++ {
		ret[i] = byte(rno.IntN(256))
	}
	return ret
}
