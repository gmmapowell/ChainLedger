package helpers

import (
	"time"

	"github.com/gmmapowell/ChainLedger/internal/types"
)

type Clock interface {
	Time() types.Timestamp
	After(d time.Duration) <-chan types.Timestamp
}

type ClockDouble struct {
	Times  []types.Timestamp
	next   int
	afters []chan types.Timestamp
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

func (clock *ClockDouble) After(d time.Duration) <-chan types.Timestamp {
	ret := make(chan types.Timestamp)
	clock.afters = append(clock.afters, ret)
	return ret
}

type ClockLive struct {
}

func (clock *ClockLive) Time() types.Timestamp {
	gotime := time.Now().UnixMilli()
	return types.Timestamp(gotime)
}

func (clock *ClockLive) After(d time.Duration) <-chan types.Timestamp {
	ret := make(chan types.Timestamp)
	mine := time.After(d)
	go func() {
		endTime := <-mine
		ret <- types.Timestamp(endTime.UnixMilli())
		close(ret)
	}()
	return ret
}
