package types

type Signature []byte

func (s Signature) MarshalBinaryInto(buf *BinaryMarsallingBuffer) {
	MarshalByteSliceInto(buf, s)
}
