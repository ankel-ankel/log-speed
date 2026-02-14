package main

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

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

func (r *int64Ring) snapshot() (oldest, newest int64, n int) {
	if r.count == 0 {
		return 0, 0, 0
	}
	newestIdx := r.idx - 1
	if newestIdx < 0 {
		newestIdx = len(r.buf) - 1
	}
	oldestIdx := 0
	if r.count == len(r.buf) {
		oldestIdx = r.idx
	}
	return r.buf[oldestIdx], r.buf[newestIdx], r.count
}

type latencyMetrics struct {
	enabled atomic.Bool

	ingestedRecords atomic.Uint64
	lastIngestNs    atomic.Int64

	mu           sync.Mutex
	ingestRecent *int64Ring
	rankLagNs    *int64Ring
}

func newLatencyMetrics(window int) *latencyMetrics {
	if window < 16 {
		window = 16
	}
	m := &latencyMetrics{
		ingestRecent: newInt64Ring(window),
		rankLagNs:    newInt64Ring(window),
	}
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
	m.ingestedRecords.Add(1)
	m.lastIngestNs.Store(nowNs)
	m.mu.Lock()
	m.ingestRecent.add(nowNs)
	m.mu.Unlock()
}

func (m *latencyMetrics) observeTopKRefresh(now time.Time) {
	if !m.isEnabled() {
		return
	}
	if now.IsZero() {
		now = time.Now()
	}
	nowNs := now.UnixNano()
	lastIngestNs := m.lastIngestNs.Load()
	lagNs := int64(0)
	if lastIngestNs > 0 && nowNs > lastIngestNs {
		lagNs = nowNs - lastIngestNs
	}
	m.mu.Lock()
	m.rankLagNs.add(lagNs)
	m.mu.Unlock()
}

type snapshot struct {
	records        uint64
	ingestRps      uint64
	ingestSamples  int
	ingestLag      time.Duration
	rankLagP95     time.Duration
	rankLagSamples int
}

func recentRate(oldest, newest int64, n int) uint64 {
	if n <= 1 || newest <= oldest {
		return 0
	}
	dt := time.Duration(newest - oldest)
	if dt <= 0 {
		return 0
	}
	rate := float64(n-1) / dt.Seconds()
	if rate <= 0 {
		return 0
	}
	return uint64(rate + 0.5)
}

func (m *latencyMetrics) snapshot() snapshot {
	if !m.isEnabled() {
		return snapshot{}
	}

	records := m.ingestedRecords.Load()
	lastIngestNs := m.lastIngestNs.Load()

	m.mu.Lock()
	oldest, newest, ingestN := m.ingestRecent.snapshot()
	rankLagP95, rankLagN := percentile95Duration(m.rankLagNs)
	m.mu.Unlock()
	ingestLag := time.Duration(0)
	if lastIngestNs > 0 {
		nowNs := time.Now().UnixNano()
		if nowNs > lastIngestNs {
			ingestLag = time.Duration(nowNs - lastIngestNs)
		}
	}

	return snapshot{
		records:        records,
		ingestRps:      recentRate(oldest, newest, ingestN),
		ingestSamples:  ingestN,
		ingestLag:      ingestLag,
		rankLagP95:     rankLagP95,
		rankLagSamples: rankLagN,
	}
}

func percentile95Duration(r *int64Ring) (time.Duration, int) {
	if r == nil || r.count == 0 {
		return 0, 0
	}
	vals := make([]int64, 0, r.count)
	for i := 0; i < r.count; i++ {
		idx := r.idx - r.count + i
		for idx < 0 {
			idx += len(r.buf)
		}
		vals = append(vals, r.buf[idx%len(r.buf)])
	}
	sort.Slice(vals, func(i, j int) bool { return vals[i] < vals[j] })
	pos := int(0.95 * float64(len(vals)-1))
	if pos < 0 {
		pos = 0
	}
	if pos >= len(vals) {
		pos = len(vals) - 1
	}
	return time.Duration(vals[pos]), len(vals)
}
