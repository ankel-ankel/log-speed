package main

import (
	"sync/atomic"
	"time"
)

type durationRing struct {
	buf   []time.Duration
	idx   int
	count int
}

func newDurationRing(n int) *durationRing {
	if n < 1 {
		n = 1
	}
	return &durationRing{buf: make([]time.Duration, n)}
}

func (r *durationRing) add(d time.Duration) {
	if len(r.buf) == 0 {
		return
	}
	r.buf[r.idx] = d
	r.idx++
	if r.idx >= len(r.buf) {
		r.idx = 0
	}
	if r.count < len(r.buf) {
		r.count++
	}
}

type durationStats struct {
	last time.Duration
	max  time.Duration
	avg  time.Duration
	n    int
}

func (r *durationRing) snapshot() durationStats {
	if r.count == 0 {
		return durationStats{}
	}
	var sum time.Duration
	var max time.Duration
	for i := 0; i < r.count; i++ {
		d := r.buf[i]
		sum += d
		if d > max {
			max = d
		}
	}

	lastIdx := r.idx - 1
	if lastIdx < 0 {
		lastIdx = len(r.buf) - 1
	}
	last := r.buf[lastIdx]

	return durationStats{
		last: last,
		max:  max,
		avg:  sum / time.Duration(r.count),
		n:    r.count,
	}
}

type latencyMetrics struct {
	enabled atomic.Bool

	startedNs       atomic.Int64
	ingestedRecords atomic.Uint64
	firstIngestNs   atomic.Int64
	lastIngestNs    atomic.Int64

	topkAny       *durationRing
	fullRefreshes atomic.Uint64
	partRefreshes atomic.Uint64
}

func newLatencyMetrics(window int) *latencyMetrics {
	m := &latencyMetrics{
		topkAny: newDurationRing(window),
	}
	m.startedNs.Store(time.Now().UnixNano())
	return m
}

func (m *latencyMetrics) setEnabled(v bool) { m.enabled.Store(v) }
func (m *latencyMetrics) isEnabled() bool   { return m.enabled.Load() }

func (m *latencyMetrics) observeIngest(now time.Time) {
	if !m.isEnabled() {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	nowNs := now.UnixNano()
	m.firstIngestNs.CompareAndSwap(0, nowNs)
	m.lastIngestNs.Store(nowNs)
	m.ingestedRecords.Add(1)
}

func (m *latencyMetrics) observeTopKRefresh(d time.Duration, didFull bool) {
	if !m.isEnabled() {
		return
	}
	m.topkAny.add(d)
	if didFull {
		m.fullRefreshes.Add(1)
		return
	}
	m.partRefreshes.Add(1)
}

type snapshot struct {
	started       time.Time
	records       uint64
	avgRps        uint64
	fullRefreshes uint64
	partRefreshes uint64
	topkLatency   durationStats
}

func (m *latencyMetrics) snapshot() snapshot {
	if !m.isEnabled() {
		return snapshot{}
	}
	startedNs := m.startedNs.Load()
	started := time.Time{}
	if startedNs != 0 {
		started = time.Unix(0, startedNs)
	}

	records := m.ingestedRecords.Load()

	lastIngestNs := m.lastIngestNs.Load()

	avgRps := uint64(0)
	firstIngestNs := m.firstIngestNs.Load()
	if firstIngestNs != 0 && lastIngestNs != 0 && lastIngestNs > firstIngestNs {
		active := time.Duration(lastIngestNs - firstIngestNs)
		if active > 0 {
			avg := float64(records) / active.Seconds()
			if avg < 0 {
				avg = 0
			}
			avgRps = uint64(avg + 0.5)
		}
	}
	return snapshot{
		started:       started,
		records:       records,
		avgRps:        avgRps,
		fullRefreshes: m.fullRefreshes.Load(),
		partRefreshes: m.partRefreshes.Load(),
		topkLatency:   m.topkAny.snapshot(),
	}
}
