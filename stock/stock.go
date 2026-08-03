// Package stock tracks item quantities.
package stock

import "errors"

// ErrUnknownItem is returned when an item has never been added.
var ErrUnknownItem = errors.New("unknown item")

type entry struct {
	name string
	qty  int
}

// Stock holds the quantitiy of each item.
type Stock struct {
	entries []entry
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

func (s *Stock) Remove(name string, qty int) error {
	for i := range s.entries {
		if s.entries[i].name == name {
			s.entries[i].qty -= qty
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
