package types

import (
	"bytes"
	"encoding/binary"
)

type BinaryMarsallingBuffer struct {
	Buf *bytes.Buffer
}

func NewBinaryMarshallingBuffer() *BinaryMarsallingBuffer {
	buf := bytes.Buffer{}
	return &BinaryMarsallingBuffer{Buf: &buf}
}

func (bmb *BinaryMarsallingBuffer) Bytes() []byte {
	return bmb.Buf.Bytes()
}

func MarshalStringInto(buf *BinaryMarsallingBuffer, s string) {
	bs := []byte(s)
	MarshalByteSliceInto(buf, bs)
}

func MarshalByteSliceInto(buf *BinaryMarsallingBuffer, bs []byte) {
	MarshalInt32Into(buf, int32(len(bs)))
	buf.Buf.Write(bs)
}

func MarshalInt32Into(buf *BinaryMarsallingBuffer, n int32) {
	binary.Write(buf.Buf, binary.LittleEndian, n)
}

func MarshalInt64Into(buf *BinaryMarsallingBuffer, n int64) {
	binary.Write(buf.Buf, binary.LittleEndian, n)
}
