package client

import (
	"crypto/rsa"
	"net/http"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
)

type Submitter struct {
	node *url.URL
	iam  *url.URL
	pk   *rsa.PrivateKey
}

func NewSubmitter(node string, id string, pk *rsa.PrivateKey) (*Submitter, error) {
	nodeAddr, e1 := url.Parse(node)
	if e1 != nil {
		return nil, e1
	}
	iam, err := url.Parse(id)
	if err != nil {
		return nil, err
	}
	return &Submitter{node: nodeAddr, iam: iam, pk: pk}, nil
}

func (s *Submitter) Submit(tx *api.Transaction) error {
	var e error = tx.Signer(s.iam)
	if e != nil {
		return e
	}
	e = tx.Sign(s.iam, s.pk)
	if e != nil {
		return e
	}
	json, e2 := tx.JsonReader()
	if e2 != nil {
		return e2
	}
	cli := http.Client{}
	_, e3 := cli.Post(s.node.JoinPath("/store").String(), "application/json", json)
	return e3
}
