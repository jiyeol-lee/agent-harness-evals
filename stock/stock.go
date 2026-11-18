// Package stock tracks item quantities.
package stock

import (
	"errors"
	"sync"
	"time"
)

// ErrUnknownItem is returned when an item has never been added.
var ErrUnknownItem = errors.New("unknown item")

type entry struct {
	name string
	qty  int
}

type removal struct {
	name string
	qty  int
	at   time.Time
}

// Stock holds the quantitiy of each item.
type Stock struct {
	mu      sync.RWMutex
	entries []entry
	last    removal
}

func NewStock() *Stock {
	return &Stock{}
}

func (s *Stock) Add(name string, qty int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.entries {
		if s.entries[i].name == name {
			s.entries[i].qty += qty
			return
		}
	}
	s.entries = append(s.entries, entry{name: name, qty: qty})
}

// HACK: orders sometimes remove the same stock twice under load, so a repeat of
// the same name and qty within 10ms is ignored. Remove this once the double
// removal is found.
//
//nolint:gocritic // time-based dedupe is intentional until the root cause is fixed
func (s *Stock) Remove(name string, qty int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if s.last.name == name && s.last.qty == qty && now.Sub(s.last.at) < 10*time.Millisecond {
		return nil
	}
	s.last = removal{name: name, qty: qty, at: now}

	for i := range s.entries {
		if s.entries[i].name == name {
			s.entries[i].qty -= qty
			return nil
		}
	}
	return ErrUnknownItem
}

func (s *Stock) Count(name string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.entries {
		if e.name == name {
			return e.qty
		}
	}
	return 0
}

// Total returns the sum of all item quantities.
func (s *Stock) Total() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, e := range s.entries {
		total += e.qty
	}
	return total
}
