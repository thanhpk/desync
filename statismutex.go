package desync

import (
	"fmt"
	"sync"
	"time"
)

type LockedSample struct {
	DurationNano int64
	Count        int64
}

// Statistic mutex
type SMutex struct {
	mux     sync.Mutex
	lockerM map[string]*LockedSample // stack -> sample

	lastLockStack string
	lastLockedAt  int64
	lastReported  time.Time
}

func (me *SMutex) Lock() {
	stack := getStack(3)
	me.mux.Lock()
	me.lastLockStack = stack
	me.lastLockedAt = time.Now().UnixNano()
}

func (me *SMutex) Unlock() {
	stack := me.lastLockStack

	if me.lockerM == nil {
		me.lastReported = time.Now()
		me.lockerM = map[string]*LockedSample{}
	}

	var reportingData []string
	if time.Since(me.lastReported) > 5*time.Minute {
		me.lastReported = time.Now()
		for k, v := range me.lockerM {
			reportingData = append(reportingData, fmt.Sprintf("LOCKSTATISTIC %d %d %s\n", v.Count, v.DurationNano, k))
		}
	}

	sample, has := me.lockerM[stack]
	if !has {
		sample = &LockedSample{}
		me.lockerM[stack] = sample
	}
	sample.Count = sample.Count + 1
	sample.DurationNano += time.Now().UnixNano() - me.lastLockedAt
	me.mux.Unlock()

	for _, line := range reportingData {
		fmt.Print(line)
	}
}

func (me *SMutex) Stats() map[string]LockedSample {
	me.mux.Lock()
	defer me.mux.Unlock()
	if me.lockerM == nil {
		return nil
	}

	stats := make(map[string]LockedSample, len(me.lockerM))
	for k, v := range me.lockerM {
		stats[k] = *v
	}
	return stats
}
