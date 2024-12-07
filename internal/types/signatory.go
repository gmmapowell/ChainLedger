package types

import (
	"encoding/json"
	"net/url"
)

type Signatory struct {
	Signer    *url.URL
	Signature *Signature
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
