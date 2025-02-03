package types

import (
	"encoding/json"
	"net/url"
)

type Signatory struct {
	Signer    *url.URL
	Signature Signature
}

func OtherSignerURL(u *url.URL) (*Signatory, error) {
	return &Signatory{Signer: u}, nil
}

func OtherSignerId(id string) (*Signatory, error) {
	u, err := url.Parse(id)
	if err != nil {
		return nil, err
	}
	return OtherSignerURL(u)
}

func (sig Signatory) MarshalJSON() ([]byte, error) {
	var m = make(map[string]any)
	m["Signer"] = sig.Signer.String()
	m["Signature"] = sig.Signature
	return json.Marshal(m)
}

func (sig *Signatory) UnmarshalJSON(bs []byte) error {
	var wire struct {
		Signer    string
		Signature Signature
	}
	if err := json.Unmarshal(bs, &wire); err != nil {
		return err
	}
	if url, err := url.Parse(wire.Signer); err == nil {
		sig.Signer = url
	} else {
		return err
	}
	sig.Signature = wire.Signature

	return nil
}

func (sg Signatory) MarshalBinaryInto(buf *BinaryMarshallingBuffer) {
	MarshalStringInto(buf, sg.Signer.String())
	sg.Signature.MarshalBinaryInto(buf)
}

func UnmarshalSignatoryFrom(buf *BinaryUnmarshallingBuffer) (*Signatory, error) {
	u, err := UnmarshalStringFrom(buf)
	if err != nil {
		return nil, err
	}
	s, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	sig, err := UnmarshalSignatureFrom(buf)
	if err != nil {
		return nil, err
	}
	return &Signatory{Signer: s, Signature: sig}, nil
}
