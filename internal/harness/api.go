package harness

import (
	"crypto/sha512"
	"encoding/json"
	"io"
	"log"
	"net/url"
	"os"
	"slices"
	"sync"
	"time"

	rno "math/rand/v2"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
	"github.com/gmmapowell/ChainLedger/internal/types"
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
	content    string
	hash       types.Hash
	originator url.URL
	cosigners  []url.URL
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
				tx, err := makeTransaction(ps, cli.user)
				if err != nil {
					log.Fatal(err)
					continue
				}
				err = cli.submitter.Submit(tx)
				if err != nil {
					log.Fatal(err)
					continue
				}
			}
		}()
	}

	go func() {
		// Publish all of "our" messages
		for i := 0; i < cli.count; i++ {
			ps, err := makeMessage(cli)
			if err != nil {
				log.Fatal(err)
				continue
			}
			tx, err := makeTransaction(ps, cli.user)
			if err != nil {
				log.Fatal(err)
				continue
			}
			err = cli.submitter.Submit(tx)
			if err != nil {
				log.Fatal(err)
				continue
			}
			for _, u := range ps.cosigners {
				cli.cosigners[u.String()] <- ps
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

// Create a random message and return it as a PleaseSign
func makeMessage(cli *ConfigClient) (PleaseSign, error) {
	content := "http://tx.info/" + randomPath()

	hasher := sha512.New()
	hasher.Write(randomBytes(16))
	h := hasher.Sum(nil)

	return PleaseSign{
		content:    content,
		hash:       h,
		originator: cli.repo.URLFor(cli.user),
		cosigners:  cli.repo.OtherThan(cli.user),
	}, nil
}

// Create a transaction from a PleaseSign request
func makeTransaction(ps PleaseSign, submitter string) (*api.Transaction, error) {
	tx, err := api.NewTransaction(ps.content, ps.hash)
	if err != nil {
		return nil, err
	}
	for _, s := range ps.cosigners {
		if s.String() == submitter {
			continue
		}
		err = tx.Signer(&s)
		if err != nil {
			return nil, err
		}
	}
	if ps.originator.String() != submitter {
		err = tx.Signer(&ps.originator)
		if err != nil {
			return nil, err
		}
	}

	return tx, nil
}

// Wait for the Begin goroutine to signal that it is fully done
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
			if repo.HasUser(cli.Client) {
				continue
			} else if err := repo.NewUser(cli.Client); err != nil {
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
			if s, err := repo.SubmitterFor(node, cli.Client); err != nil {
				panic(err)
			} else {
				client := ConfigClient{
					repo:      &repo,
					submitter: s,
					user:      cli.Client,
					count:     cli.Count,
					signFor:   chanReceivers(chans, cli.Client),
					cosigners: chanSenders(chans, cli.Client),
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
		if slices.Index(ret, c.Client) == -1 {
			ret = append(ret, c.Client)
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
