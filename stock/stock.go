// Package stock tracks item quantities.
package stock

import (
	"errors"
	"time"
)

// ErrUnknownItem is returned when an item has never been added.
var ErrUnknownItem = errors.New("unknown item")

// ErrInsufficientStock is returned when a removal exceeds the available quantity.
var ErrInsufficientStock = errors.New("insufficient stock")

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
	entries []entry
	last    removal
}

func NewStock() *Stock {
	return &Stock{}
}

func (s *Stock) Add(name string, qty int) {
	for i := range s.entries {
		if s.entries[i].name == name {
			s.entries[i].qty += qty
			return
		}
	}
	s.entries = append(s.entries, entry{name: name, qty: qty})
}

// HACK: orders sometimes remove the same stock twice under load, so a repeat of
// the same name and qty within 10ms is ignored if it fits the available stock.
// Remove this once the double removal is found.
//
//nolint:gocritic // time-based dedupe is intentional until the root cause is fixed
func (s *Stock) Remove(name string, qty int) error {
	for i := range s.entries {
		if s.entries[i].name == name {
			if qty > s.entries[i].qty {
				return ErrInsufficientStock
			}
			now := time.Now()
			if s.last.name == name && s.last.qty == qty && now.Sub(s.last.at) < 10*time.Millisecond {
				return nil
			}
			s.entries[i].qty -= qty
			s.last = removal{name: name, qty: qty, at: now}
			return nil
		}
	}
	return ErrUnknownItem
}

func (s *Stock) Count(name string) int {
	for _, e := range s.entries {
		if e.name == name {
			return e.qty
		}
	}
	return 0
}
