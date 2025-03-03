package records

import (
	"bytes"
	"crypto/rsa"
	"encoding/binary"
	"fmt"
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

func (s *StoredTransaction) VerifySignature(hasher helpers.HasherFactory, signer helpers.Signer, pub *rsa.PublicKey) error {
	txid := s.hashMe(hasher)
	if !txid.Is(s.TxID) {
		return fmt.Errorf("remote txid %s was not the result of computing it locally: %s", s.TxID.String(), txid.String())
	}
	return signer.Verify(pub, txid, s.Publisher.Signature)
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
	for i, v := range tx.Signatories {
		copySigner := *v.Signer
		copySig := types.Signature(bytes.Clone(v.Signature))
		signatory := types.Signatory{Signer: &copySigner, Signature: copySig}
		ret.Signatories[i] = &signatory
	}
	ret.TxID = ret.hashMe(hasherFactory)

	sig, err := signer.Sign(nodeKey, ret.TxID)
	if err != nil {
		return nil, err
	}
	ret.Publisher = &types.Signatory{Signer: signer.SignerName(), Signature: sig}

	return &ret, nil
}

func (stx *StoredTransaction) hashMe(hasherFactory helpers.HasherFactory) types.Hash {
	hasher := hasherFactory.NewHasher()
	binary.Write(hasher, binary.LittleEndian, stx.WhenReceived)
	hasher.Write([]byte(stx.ContentLink.String()))
	hasher.Write([]byte("\n"))
	hasher.Write(stx.ContentHash)
	for _, v := range stx.Signatories {
		hasher.Write([]byte(v.Signer.String()))
		hasher.Write([]byte("\n"))
		hasher.Write(v.Signature)
	}
	return hasher.Sum(nil)
}
