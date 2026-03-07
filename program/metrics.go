package main

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type latencyMetrics struct {
	enabled atomic.Bool

	ingestedRecords atomic.Uint64
	lastEventTimeNs atomic.Int64

	mu              sync.Mutex
	rateCounter     int64
	rateLastBucket  int64
	rateBuckets     *int64Ring
	refreshDurationNs *int64Ring
}

func newLatencyMetrics(window int) *latencyMetrics {
	if window < 16 {
		window = 16
	}
	return &latencyMetrics{
		rateBuckets:       newInt64Ring(window),
		refreshDurationNs: newInt64Ring(window),
	}
}

func (m *latencyMetrics) setEnabled(v bool) { m.enabled.Store(v) }

func (m *latencyMetrics) observeIngest(now time.Time) {
	if !m.enabled.Load() {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	m.ingestedRecords.Add(1)

	bucket := now.Unix()
	m.mu.Lock()
	if bucket == m.rateLastBucket {
		m.rateCounter++
	} else {
		if m.rateLastBucket > 0 {
			m.rateBuckets.add(m.rateCounter)
		}
		m.rateLastBucket = bucket
		m.rateCounter = 1
	}
	m.mu.Unlock()
}

func (m *latencyMetrics) observeEventTime(t time.Time) {
	if !t.IsZero() {
		m.lastEventTimeNs.Store(t.UnixNano())
	}
}

func (m *latencyMetrics) observeRefreshDuration(d time.Duration) {
	if !m.enabled.Load() {
		return
	}
	m.mu.Lock()
	m.refreshDurationNs.add(int64(d))
	m.mu.Unlock()
}

type snapshot struct {
	records       uint64
	ingestRps     int64
	lastEventTime time.Time
	refreshP95    time.Duration
}

func (m *latencyMetrics) snapshot() snapshot {
	if !m.enabled.Load() {
		return snapshot{}
	}

	records := m.ingestedRecords.Load()
	lastEventNs := m.lastEventTimeNs.Load()

	m.mu.Lock()
	rps := medianRate(m.rateBuckets)
	refreshP95, _ := percentile95(m.refreshDurationNs)
	m.mu.Unlock()

	var lastEventTime time.Time
	if lastEventNs > 0 {
		lastEventTime = time.Unix(0, lastEventNs)
	}

	return snapshot{
		records:       records,
		ingestRps:     rps,
		lastEventTime: lastEventTime,
		refreshP95:    refreshP95,
	}
}

// medianRate returns the median of recent per-second record counts.
func medianRate(r *int64Ring) int64 {
	if r == nil || r.count == 0 {
		return 0
	}
	vals := ringValues(r)
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	return vals[len(vals)/2]
}

func percentile95(r *int64Ring) (time.Duration, int) {
	if r == nil || r.count == 0 {
		return 0, 0
	}
	vals := ringValues(r)
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	pos := int(0.95 * float64(len(vals)-1))
	if pos >= len(vals) {
		pos = len(vals) - 1
	}
	return time.Duration(vals[pos]), len(vals)
}

func ringValues(r *int64Ring) []int64 {
	vals := make([]int64, 0, r.count)
	for i := 0; i < r.count; i++ {
		idx := r.idx - r.count + i
		for idx < 0 {
			idx += len(r.buf)
		}
		vals = append(vals, r.buf[idx%len(r.buf)])
	}
	return vals
}

// int64Ring is a fixed-size circular buffer of int64 values.
type int64Ring struct {
	buf   []int64
	idx   int
	count int
}

func newInt64Ring(n int) *int64Ring {
	if n < 1 {
		n = 1
	}
	return &int64Ring{buf: make([]int64, n)}
}

func (r *int64Ring) add(v int64) {
	if len(r.buf) == 0 {
		return
	}
	r.buf[r.idx] = v
	r.idx++
	if r.idx >= len(r.buf) {
		r.idx = 0
	}
	if r.count < len(r.buf) {
		r.count++
	}
}
