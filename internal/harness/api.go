package harness

import (
	"slices"

	"github.com/gmmapowell/ChainLedger/internal/client"
	"github.com/gmmapowell/ChainLedger/internal/clienthandler"
)

// Start the nodes
func StartNodes(c Config) []clienthandler.Node {
	var nodes []clienthandler.Node
	for _, n := range c.NodeNames() {
		node := clienthandler.NewListenerNode(c.Launcher(n))
		nodes = append(nodes, node)
		go node.Start()
	}
	return nodes
}

// Build up the list of all the clients
func PrepareClients(c Config) []Client {
	// Create the one and only client side repo
	repo, err := client.MakeMemoryRepo()
	if err != nil {
		panic(err)
	}

	// Find all the users who are connecting to nodes and make sure they are in the repo
	for _, n := range c.NodeNames() {
		clis := c.ClientsFor(n)
		for _, cli := range clis {
			if repo.HasUser(cli.User) {
				continue
			} else if err := repo.NewUser(cli.User); err != nil {
				panic(err)
			}
		}
	}

	// Create all the clients that will publish, and make sure that they also have all the corresponding listeners
	ret := make([]Client, 0)
	for _, n := range c.NodeNames() {
		clis := c.ClientsFor(n)
		// Figure out all the users on this node, and thus the cross-product of co-signing channels we need
		allUsers := usersOnNode(clis)
		chans := crossChannels(allUsers)

		// Now create the submitters and thus the clients and build a list
		for _, cli := range clis {
			if s, err := repo.SubmitterFor(n, cli.User); err != nil {
				panic(err)
			} else {
				client := ConfigClient{
					repo:      &repo,
					submitter: s,
					user:      cli.User,
					count:     cli.Count,
					signFor:   chanReceivers(chans, cli.User),
					cosigners: chanSenders(chans, cli.User),
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
func usersOnNode(clis []*CliConfig) []string {
	ret := make([]string, 0)
	for _, c := range clis {
		if slices.Index(ret, c.User) == -1 {
			ret = append(ret, c.User)
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
