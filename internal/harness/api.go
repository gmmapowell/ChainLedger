package harness

import (
	"crypto/sha512"
	"log"
	"slices"
	"sync"
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
	cosigners map[string]chan<- PleaseSign
	signFor   []<-chan PleaseSign
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

type PleaseSign struct {
}

func (cli *ConfigClient) Begin() {
	// We need to coordinate activity across all the cosigner threads, so create a waitgroup
	var wg sync.WaitGroup

	// Create one goroutine for each of the other clients attached to "this" node which might ask us to cosign for them
	for _, c := range cli.signFor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ps := range c {
				log.Printf("have message to sign: %v\n", ps)
			}
		}()
	}

	go func() {
		// Publish all of "our" messages
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

		// Tell all of the remote cosigners that we have finished by closing the channels
		for _, c := range cli.cosigners {
			close(c)
		}

		// Make sure all of our cosigners have finished
		wg.Wait()

		// Now we can report that we are fully done
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

// Start the nodes
func StartNodes(c Config) {
	for _, ep := range c.NodeEndpoints() {
		node := clienthandler.NewListenerNode(ep)
		go node.Start()
	}
}

// Build up the list of all the clients
func PrepareClients(c Config) []Client {
	// Create the one and only client side repo
	repo, err := client.MakeMemoryRepo()
	if err != nil {
		panic(err)
	}

	// Find all the users who are connecting to nodes and make sure they are in the repo
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

	// Create all the clients that will publish, and make sure that they also have all the corresponding listeners
	ret := make([]Client, 0)
	for node, clis := range m {
		// Figure out all the users on this node, and thus the cross-product of co-signing channels we need
		allUsers := usersOnNode(clis)
		chans := crossChannels(allUsers)

		// Now create the submitters and thus the clients and build a list
		for _, cli := range clis {
			if s, err := repo.SubmitterFor(node, cli.client); err != nil {
				panic(err)
			} else {
				client := ConfigClient{
					repo:      &repo,
					submitter: s,
					user:      cli.client,
					count:     cli.count,
					signFor:   chanReceivers(chans, cli.client),
					cosigners: chanSenders(chans, cli.client),
					done:      make(chan struct{}),
				}
				ret = append(ret, &client)
			}
		}
	}

	for _, s := range ret {
		s.PingNode()
	}
	return ret
}

// Find all the users in a list of clients associated with a given node
func usersOnNode(clis []CliConfig) []string {
	ret := make([]string, 0)
	for _, c := range clis {
		if slices.Index(ret, c.client) == -1 {
			ret = append(ret, c.client)
		}
	}
	return ret
}

// Create a cross-product of all the channels that request counterparties to sign
func crossChannels(allUsers []string) map[string]map[string]chan PleaseSign {
	ret := make(map[string]map[string]chan PleaseSign)
	for _, from := range allUsers {
		for _, to := range allUsers {
			if from == to {
				continue
			}
			m1, e1 := ret[from]
			if !e1 {
				m1 = make(map[string]chan PleaseSign)
				ret[from] = m1
			}
			m1[to] = make(chan PleaseSign)
		}
	}
	return ret
}

// Extract all the "from" entries for a given user as a map of user id -> sending channel
func chanSenders(chans map[string]map[string]chan PleaseSign, user string) map[string]chan<- PleaseSign {
	ret := make(map[string]chan<- PleaseSign, 0)
	for u, c := range chans[user] {
		ret[u] = c
	}
	return ret
}

// Extract all the "to" entries for a given user as receiving channels
func chanReceivers(chans map[string]map[string]chan PleaseSign, user string) []<-chan PleaseSign {
	ret := make([]<-chan PleaseSign, 0)
	for _, m := range chans {
		for u, c := range m {
			if u == user {
				ret = append(ret, c)
			}
		}
	}
	return ret
}

// Generate a random string to use as the "unique" message path
func randomPath() string {
	ns := 6 + rno.IntN(6)
	ret := make([]rune, ns)
	for i := 0; i < ns; i++ {
		ret[i] = alnumRune()
	}
	return string(ret)
}

// Generate a random character from a-z._
func alnumRune() rune {
	r := rno.IntN(38)
	switch {
	case r == 0:
		return '_'
	case r == 1:
		return '.'
	case r >= 2 && r < 12:
		return rune('0' + r - 2)
	case r >= 12:
		return rune('a' + r - 12)
	}
	panic("this should be in the range 0-38")
}

// Generate a random set of bytes to be used as a hash
func randomBytes(ns int) []byte {
	ret := make([]byte, ns)
	for i := 0; i < ns; i++ {
		ret[i] = byte(rno.IntN(256))
	}
	return ret
}
