package records

import (
	"bytes"
	"net/url"

	"github.com/gmmapowell/ChainLedger/internal/api"
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

func CreateStoredTransaction(tx *api.Transaction) *StoredTransaction {
	copyLink := *tx.ContentLink
	ret := StoredTransaction{ContentLink: &copyLink, ContentHash: bytes.Clone(tx.ContentHash), Signatories: make([]*types.Signatory, len(tx.Signatories))}
	for i, v := range tx.Signatories {
		copySigner := *v.Signer
		copySig := types.Signature(bytes.Clone(*v.Signature))
		signatory := types.Signatory{Signer: &copySigner, Signature: &copySig}
		ret.Signatories[i] = &signatory
	}
	return &ret
}
