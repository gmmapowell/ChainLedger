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
	Publisher    *types.Signatory
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
	s.Publisher.MarshalBinaryInto(ret)
	return ret.Bytes(), nil
}

func UnmarshalBinaryStoredTransaction(bytes []byte) (*StoredTransaction, error) {
	buf := types.NewBinaryUnmarshallingBuffer(bytes)
	stx := StoredTransaction{}
	var err error
	stx.TxID, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, err
	}
	stx.WhenReceived, err = types.UnmarshalTimestampFrom(buf)
	if err != nil {
		return nil, err
	}
	cls, err := types.UnmarshalStringFrom(buf)
	if err != nil {
		return nil, err
	}
	stx.ContentLink, err = url.Parse(cls)
	if err != nil {
		return nil, err
	}
	stx.ContentHash, err = types.UnmarshalHashFrom(buf)
	if err != nil {
		return nil, err
	}
	nsigs, err := types.UnmarshalInt32From(buf)
	if err != nil {
		return nil, err
	}
	stx.Signatories = make([]*types.Signatory, nsigs)
	for i := 0; i < int(nsigs); i++ {
		stx.Signatories[i], err = types.UnmarshalSignatoryFrom(buf)
		if err != nil {
			return nil, err
		}
	}
	stx.Publisher, err = types.UnmarshalSignatoryFrom(buf)
	if err != nil {
		return nil, err
	}
	err = buf.ShouldBeDone()
	if err != nil {
		return nil, err
	}
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
	ret.Publisher = &types.Signatory{Signer: signer.SignerName(), Signature: sig}

	return &ret, nil
}
