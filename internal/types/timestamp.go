package types

import (
	"encoding/binary"
	"time"
)

type Timestamp int64

const IsoFormat = "2006-01-02_15:04:05.999"

func ParseTimestamp(iso string) (Timestamp, error) {
	ts, err := time.Parse(IsoFormat, iso)
	if err != nil {
		return Timestamp(0), err
	}
	return Timestamp(ts.UnixMilli()), nil

}

func (ts Timestamp) IsoTime() string {
	return time.UnixMilli(int64(ts)).Format(IsoFormat)
}

func (ts Timestamp) AsBytes() []byte {
	var s = make([]byte, 8)
	binary.LittleEndian.PutUint64(s, uint64(ts))
	return s
}

func (ts Timestamp) MarshalBinaryInto(buf *BinaryMarshallingBuffer) {
	MarshalInt64Into(buf, int64(ts))
}

func UnmarshalTimestampFrom(buf *BinaryUnmarshallingBuffer) (Timestamp, error) {
	i64, err := buf.ReadInt64()
	if err != nil {
		return 0, err
	}
	return Timestamp(i64), nil
}

func (ts Timestamp) RoundTime(granularity int) Timestamp {
	i64 := int64(ts)
	i64 /= int64(granularity)
	i64 *= int64(granularity)
	return Timestamp(i64)
}
