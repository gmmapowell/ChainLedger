package types

import (
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
