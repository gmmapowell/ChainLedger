package types

type Hash []byte

func (h Hash) MarshalBinaryInto(buf *BinaryMarshallingBuffer) {
	MarshalByteSliceInto(buf, h)
}

func UnmarshalHashFrom(buf *BinaryUnmarshallingBuffer) (Hash, error) {
	return UnmarshalByteSliceFrom(buf)
}
