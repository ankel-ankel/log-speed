package main

import (
	"sort"
	"time"

	"github.com/keilerkonzept/topk/heap"
)

type IncrementalRanker struct {
	k           int
	fullRefresh time.Duration
	partialSize int

	lastFullRefresh time.Time
	items           []heap.Item
}

func NewIncrementalRanker(k int, fullRefresh time.Duration, partialSize int) *IncrementalRanker {
	if k < 1 {
		k = 1
	}
	if fullRefresh < 0 {
		fullRefresh = 2 * time.Second
	}
	if partialSize < 0 {
		partialSize = 0
	}
	return &IncrementalRanker{
		k:           k,
		fullRefresh: fullRefresh,
		partialSize: partialSize,
	}
}

// Refresh updates the view of the top-K items.
// Provide two functions that are safe to call while holding any shared locks:
// - sortedFn: returns a full Top-K view (approximate, from the sketch)
// - updateCountsFn: updates Count fields for the first `limit` items in-place
func (r *IncrementalRanker) Refresh(now time.Time, visibleItems int, sortedFn func() []heap.Item, updateCountsFn func(items []heap.Item, limit int)) (items []heap.Item, didFull bool) {
	if now.IsZero() {
		now = time.Now()
	}

	needFull := len(r.items) == 0 || r.lastFullRefresh.IsZero()
	if r.fullRefresh == 0 {
		needFull = true
	} else if r.fullRefresh > 0 && now.Sub(r.lastFullRefresh) >= r.fullRefresh {
		needFull = true
	}
	if needFull {
		r.items = sortedFn()
		if len(r.items) > r.k {
			r.items = r.items[:r.k]
		}
		r.lastFullRefresh = now
		return cloneItems(r.items), true
	}

	limit := len(r.items)
	if visibleItems > 0 && visibleItems < limit {
		limit = visibleItems
	}
	if r.partialSize > 0 && r.partialSize < limit {
		limit = r.partialSize
	}

	updateCountsFn(r.items, limit)

	sort.SliceStable(r.items[:limit], func(i, j int) bool {
		li := r.items[i]
		lj := r.items[j]
		if li.Count != lj.Count {
			return li.Count > lj.Count
		}
		return li.Item < lj.Item
	})

	return cloneItems(r.items), false
}

func cloneItems(in []heap.Item) []heap.Item {
	out := make([]heap.Item, len(in))
	copy(out, in)
	return out
}
