package types

type Signature []byte

func (s Signature) MarshalBinaryInto(buf *BinaryMarshallingBuffer) {
	MarshalByteSliceInto(buf, s)
}

func UnmarshalSignatureFrom(buf *BinaryUnmarshallingBuffer) (Signature, error) {
	return UnmarshalByteSliceFrom(buf)
}
