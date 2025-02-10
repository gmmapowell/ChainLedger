package types

import (
	"bytes"
	"encoding/binary"
)

type BinaryMarshallingBuffer struct {
	Buf *bytes.Buffer
}

func NewBinaryMarshallingBuffer() *BinaryMarshallingBuffer {
	buf := bytes.Buffer{}
	return &BinaryMarshallingBuffer{Buf: &buf}
}

func (bmb *BinaryMarshallingBuffer) Bytes() []byte {
	return bmb.Buf.Bytes()
}

func MarshalStringInto(buf *BinaryMarshallingBuffer, s string) {
	bs := []byte(s)
	MarshalByteSliceInto(buf, bs)
}

func UnmarshalStringFrom(buf *BinaryUnmarshallingBuffer) (string, error) {
	bs, err := UnmarshalByteSliceFrom(buf)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func MarshalByteSliceInto(buf *BinaryMarshallingBuffer, bs []byte) {
	MarshalInt32Into(buf, int32(len(bs)))
	buf.Buf.Write(bs)
}

func UnmarshalByteSliceFrom(buf *BinaryUnmarshallingBuffer) ([]byte, error) {
	
	ilen, err := buf.ReadInt32()
	if err != nil {
		return nil, err
	}
	return buf.ReadBytes(int(ilen))
}

func MarshalInt32Into(buf *BinaryMarshallingBuffer, n int32) {
	binary.Write(buf.Buf, binary.LittleEndian, n)
}

func UnmarshalInt32From(buf *BinaryUnmarshallingBuffer) (int32, error) {
	i32, err := buf.ReadInt32()
	if err != nil {
		return 0, err
	}
	return i32, nil
}

func MarshalInt64Into(buf *BinaryMarshallingBuffer, n int64) {
	binary.Write(buf.Buf, binary.LittleEndian, n)
}
