package types

import (
	"net/url"
)

type Signatory struct {
	signer    url.URL
	signature Signature
}
