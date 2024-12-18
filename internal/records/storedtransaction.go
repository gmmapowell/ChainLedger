package records

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/binary"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
	"github.com/gmmapowell/ChainLedger/internal/helpers"
	"github.com/gmmapowell/ChainLedger/internal/types"
)

type StoredTransaction struct {
	TxID         types.Hash
	WhenReceived types.Timestamp
	ContentLink  *url.URL
	ContentHash  types.Hash
	Signatories  []*types.Signatory
	NodeSig      *types.Signature
}

func CreateStoredTransaction(clock helpers.Clock, nodeKey *rsa.PrivateKey, tx *api.Transaction) (*StoredTransaction, error) {
	copyLink := *tx.ContentLink
	ret := StoredTransaction{WhenReceived: clock.Time(), ContentLink: &copyLink, ContentHash: bytes.Clone(tx.ContentHash), Signatories: make([]*types.Signatory, len(tx.Signatories))}
	hasher := sha512.New()
	binary.Write(hasher, binary.LittleEndian, ret.WhenReceived)
	hasher.Write([]byte(ret.ContentLink.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(tx.ContentHash)
	for i, v := range tx.Signatories {
		copySigner := *v.Signer
		hasher.Write([]byte(copySigner.String()))
		hasher.Write([]byte("\n"))
		copySig := types.Signature(bytes.Clone(*v.Signature))
		hasher.Write(copySig)
		signatory := types.Signatory{Signer: &copySigner, Signature: &copySig}
		ret.Signatories[i] = &signatory
	}
	ret.TxID = hasher.Sum(nil)

	sig, err := rsa.SignPSS(rand.Reader, nodeKey, crypto.SHA512, ret.TxID, nil)
	if err != nil {
		return nil, err
	}
	sig1 := types.Signature(sig)
	ret.NodeSig = &sig1

	return &ret, nil
}
