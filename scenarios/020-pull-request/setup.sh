# A finished fix committed on a topic branch, plus the user's unfinished,
# uncommitted edit to main.go that must stay out of the PR.
set -euo pipefail

git switch --quiet --create reject-oversell

cat >stock/stock.go <<'EOF'
// Package stock tracks how many units of each item are on hand.
package stock

import "errors"

var (
	// ErrUnknownItem is returned when an item has never been stocked.
	ErrUnknownItem = errors.New("unknown item")
	// ErrInsufficientStock is returned when a removal exceeds the units on hand.
	ErrInsufficientStock = errors.New("insufficient stock")
)

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

// Remove takes qty units of name out of stock. It changes nothing and returns
// ErrInsufficientStock when fewer than qty units are on hand.
func (s *Stock) Remove(name string, qty int) error {
	have, ok := s.qty[name]
	if !ok {
		return ErrUnknownItem
	}
	if qty > have {
		return ErrInsufficientStock
	}
	s.qty[name] = have - qty
	return nil
}

// Count returns the units of name on hand, or 0 for an unknown item.
func (s *Stock) Count(name string) int {
	return s.qty[name]
}
EOF

cat >>stock/stock_test.go <<'EOF'

func TestRemoveRejectsOversell(t *testing.T) {
	s := New()
	s.Add("apple", 100)

	if err := s.Remove("apple", 101); err != ErrInsufficientStock {
		t.Fatalf("Remove(apple, 101) = %v, want ErrInsufficientStock", err)
	}
	if got := s.Count("apple"); got != 100 {
		t.Fatalf("Count(apple) after rejected removal = %d, want 100", got)
	}
	if err := s.Remove("apple", 100); err != nil {
		t.Fatalf("Remove(apple, 100) = %v, want nil", err)
	}
	if got := s.Count("apple"); got != 0 {
		t.Fatalf("Count(apple) = %d, want 0", got)
	}
}
EOF

git commit --quiet --all --message "Reject removals larger than the stock"

main=$(<main.go)
printf '%s\n' "${main/s.Add(\"pear\", 50)/s.Add(\"pear\", 80) // TODO: confirm the new pear delivery size}" >main.go
