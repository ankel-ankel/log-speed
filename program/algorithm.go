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
	autoBudget  int

	lastFullRefresh time.Time
	items           []heap.Item
	partialCursor   int
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
	// Auto mode refreshes about half of Top-K each tick.
	autoBudget := k / 2
	if autoBudget < 1 {
		autoBudget = 1
	}
	if k >= 10 && autoBudget < 10 {
		autoBudget = 10
	}
	if autoBudget > 100 {
		autoBudget = 100
	}
	if autoBudget > k {
		autoBudget = k
	}
	return &IncrementalRanker{
		k:           k,
		fullRefresh: fullRefresh,
		partialSize: partialSize,
		autoBudget:  autoBudget,
	}
}

func (r *IncrementalRanker) Refresh(now time.Time, budgetItems int, sortedFn func() []heap.Item, updateCountsFn func(items []heap.Item, limit int)) (items []heap.Item, didFull bool) {
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
		discovered := sortedFn()
		if len(discovered) > r.k {
			discovered = discovered[:r.k]
		}
		r.items = cloneItems(discovered)
		r.partialCursor = 0
		r.lastFullRefresh = now
		if len(r.items) == 0 {
			return nil, true
		}
		return cloneItems(r.items), true
	}
	if len(r.items) == 0 {
		return nil, needFull
	}

	limit := len(r.items)
	if r.partialSize > 0 {
		if r.partialSize < limit {
			limit = r.partialSize
		}
	} else {
		if budgetItems <= 0 {
			budgetItems = r.autoBudget
		}
		if budgetItems > 0 && budgetItems < limit {
			limit = budgetItems
		}
	}
	if limit <= 0 || len(r.items) == 0 {
		return cloneItems(r.items), needFull
	}

	if limit >= len(r.items) {
		updateCountsFn(r.items, len(r.items))
	} else {
		start := r.partialCursor % len(r.items)
		end := start + limit
		if end <= len(r.items) {
			seg := r.items[start:end]
			updateCountsFn(seg, len(seg))
		} else {
			segA := r.items[start:]
			segB := r.items[:end-len(r.items)]
			updateCountsFn(segA, len(segA))
			updateCountsFn(segB, len(segB))
		}
		r.partialCursor = (start + limit) % len(r.items)
	}

	sort.SliceStable(r.items, func(i, j int) bool {
		li := r.items[i]
		lj := r.items[j]
		if li.Count != lj.Count {
			return li.Count > lj.Count
		}
		return li.Item < lj.Item
	})

	end := len(r.items)
	for end > 0 && r.items[end-1].Count == 0 {
		end--
	}
	r.items = r.items[:end]
	if len(r.items) == 0 {
		r.partialCursor = 0
	}

	return cloneItems(r.items), needFull
}

func cloneItems(in []heap.Item) []heap.Item {
	out := make([]heap.Item, len(in))
	copy(out, in)
	return out
}
