package helpers

import (
	"time"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Clock interface {
	Time() types.Timestamp
}

type ClockDouble struct {
	Times []types.Timestamp
	next  int
}

func ClockDoubleIsoTimes(isoTimes ...string) ClockDouble {
	ts := make([]types.Timestamp, len(isoTimes))
	for i, s := range isoTimes {
		ts[i], _ = types.ParseTimestamp(s)
	}
	return ClockDouble{Times: ts, next: 0}
}

func ClockDoubleSameDay(isoDate string, times ...string) ClockDouble {
	ts := make([]types.Timestamp, len(times))
	for i, s := range times {
		ts[i], _ = types.ParseTimestamp(isoDate + "_" + s)
	}
	return ClockDouble{Times: ts, next: 0}
}

func ClockDoubleSameMinute(isoDateHM string, seconds ...string) ClockDouble {
	ts := make([]types.Timestamp, len(seconds))
	for i, s := range seconds {
		ts[i], _ = types.ParseTimestamp(isoDateHM + ":" + s)
	}
	return ClockDouble{Times: ts, next: 0}
}

func (clock *ClockDouble) Time() types.Timestamp {
	if clock.next > len(clock.Times) {
		panic("more timestamps requested than provided")
	}
	r := clock.Times[clock.next]
	clock.next++
	return r
}

type ClockLive struct {
}

func (clock *ClockLive) Time() types.Timestamp {
	gotime := time.Now().UnixMilli()
	return types.Timestamp(gotime)
}
