# A topic branch whose refactor of Remove forgets to store the new count, so
# the checked-in TestRemove fails. The runner opens a draft PR for it.
set -euo pipefail

git switch --quiet --create "remove-lookup-$TAG"

cat >stock/stock.go <<'EOF'
// Package stock tracks how many units of each item are on hand.
package stock

import "errors"

// ErrUnknownItem is returned when an item has never been stocked.
var ErrUnknownItem = errors.New("unknown item")

// Stock maps item names to the units on hand.
type Stock struct {
	qty map[string]int
}

// New returns an empty Stock.
func New() *Stock {
	return &Stock{qty: map[string]int{}}
}

// Add puts qty units of name into stock.
func (s *Stock) Add(name string, qty int) {
	s.qty[name] += qty
}

// Remove takes qty units of name out of stock.
func (s *Stock) Remove(name string, qty int) error {
	left, ok := s.qty[name]
	if !ok {
		return ErrUnknownItem
	}
	left -= qty
	return nil
}

// Count returns the units of name on hand, or 0 for an unknown item.
func (s *Stock) Count(name string) int {
	return s.qty[name]
}
EOF

git commit --quiet --all --message "Look up the item once in Remove"
