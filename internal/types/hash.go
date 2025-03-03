package types

import (
	"bytes"
	"encoding/hex"
)

type Hash []byte

func (h Hash) Is(other Hash) bool {
	return bytes.Equal(h, other)
}

func (h Hash) MarshalBinaryInto(buf *BinaryMarshallingBuffer) {
	MarshalByteSliceInto(buf, h)
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func UnmarshalHashFrom(buf *BinaryUnmarshallingBuffer) (Hash, error) {
	return UnmarshalByteSliceFrom(buf)
}
