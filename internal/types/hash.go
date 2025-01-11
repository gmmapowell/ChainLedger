package types

type Hash []byte

func (h Hash) MarshalBinaryInto(buf *BinaryMarsallingBuffer) {
	MarshalByteSliceInto(buf, h)
}
