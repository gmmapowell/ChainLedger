package records

import (
	"bytes"
	"crypto/rsa"
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
	NodeSig      types.Signature
}

func (s *StoredTransaction) MarshalBinary() ([]byte, error) {
	ret := types.NewBinaryMarshallingBuffer()
	s.TxID.MarshalBinaryInto(ret)
	s.WhenReceived.MarshalBinaryInto(ret)
	types.MarshalStringInto(ret, s.ContentLink.String())
	s.ContentHash.MarshalBinaryInto(ret)
	types.MarshalInt32Into(ret, int32(len(s.Signatories)))
	for _, sg := range s.Signatories {
		sg.MarshalBinaryInto(ret)
	}
	s.NodeSig.MarshalBinaryInto(ret)
	return ret.Bytes(), nil
}

func UnmarshalBinaryStoredTransaction(bytes []byte) (*StoredTransaction, error) {
	buf := types.NewBinaryUnmarshallingBuffer(bytes)
	stx := StoredTransaction{}
	stx.TxID, _ = types.UnmarshalHashFrom(buf)
	stx.WhenReceived, _ = types.UnmarshalTimestampFrom(buf)
	cls, _ := types.UnmarshalStringFrom(buf)
	stx.ContentLink, _ = url.Parse(cls)
	stx.ContentHash, _ = types.UnmarshalHashFrom(buf)
	nsigs, _ := types.UnmarshalInt32From(buf)
	stx.Signatories = make([]*types.Signatory, nsigs)
	for i := 0; i < int(nsigs); i++ {
		stx.Signatories[i], _ = types.UnmarshalSignatoryFrom(buf)
	}
	_ = buf.ShouldBeDone()
	return &stx, nil
}

func CreateStoredTransaction(clock helpers.Clock, hasherFactory helpers.HasherFactory, signer helpers.Signer, nodeKey *rsa.PrivateKey, tx *api.Transaction) (*StoredTransaction, error) {
	copyLink := *tx.ContentLink
	ret := StoredTransaction{WhenReceived: clock.Time(), ContentLink: &copyLink, ContentHash: bytes.Clone(tx.ContentHash), Signatories: make([]*types.Signatory, len(tx.Signatories))}
	hasher := hasherFactory.NewHasher()
	binary.Write(hasher, binary.LittleEndian, ret.WhenReceived)
	hasher.Write([]byte(ret.ContentLink.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(tx.ContentHash)
	for i, v := range tx.Signatories {
		copySigner := *v.Signer
		hasher.Write([]byte(copySigner.String()))
		hasher.Write([]byte("\n"))
		copySig := types.Signature(bytes.Clone(v.Signature))
		hasher.Write(copySig)
		signatory := types.Signatory{Signer: &copySigner, Signature: copySig}
		ret.Signatories[i] = &signatory
	}
	ret.TxID = hasher.Sum(nil)

	sig, err := signer.Sign(nodeKey, ret.TxID)
	if err != nil {
		return nil, err
	}
	ret.NodeSig = types.Signature(sig)

	return &ret, nil
}
