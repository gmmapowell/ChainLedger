package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type BinaryUnmarshallingBuffer struct {
	buf *bytes.Buffer
}

func (b *BinaryUnmarshallingBuffer) ShouldBeDone() error {
	if b.buf.Available() > 0 {
		return fmt.Errorf("there were still bytes in the buffer at the end")
	}
	return nil
}

func (b *BinaryUnmarshallingBuffer) ReadBytes(ilen int) ([]byte, error) {
	bs := make([]byte, ilen)
	n, err := b.buf.Read(bs)
	if err != nil {
		return nil, err
	} else if n != ilen {
		return nil, fmt.Errorf("insufficient bytes in buffer for string")
	}
	return bs, nil
}

func (b *BinaryUnmarshallingBuffer) ReadInt32() (int32, error) {
	p := make([]byte, 4)
	n, err := b.buf.Read(p)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("insufficient bytes remaining")
	}
	return int32(binary.LittleEndian.Uint32(p)), nil
}

func (b *BinaryUnmarshallingBuffer) ReadInt64() (int64, error) {
	p := make([]byte, 8)
	n, err := b.buf.Read(p)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("insufficient bytes remaining")
	}
	return int64(binary.LittleEndian.Uint64(p)), nil
}

func NewBinaryUnmarshallingBuffer(bs []byte) *BinaryUnmarshallingBuffer {
	return &BinaryUnmarshallingBuffer{buf: bytes.NewBuffer(bs)}
}
