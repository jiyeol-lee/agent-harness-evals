package stock

import (
	"fmt"
	"math/big"
	"testing"
)

func TestAddAndCount(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("apple", 2)

	if got := s.Count("apple"); got != 5 {
		t.Fatalf("Count(apple) = %d, want 5", got)
	}
}

func TestRemoveUnknown(t *testing.T) {
	s := NewStock()

	if err := s.Remove("pear", 1); err != ErrUnknownItem {
		t.Fatalf("Remove(pear) = %v, want ErrUnknownItem", err)
	}
}

func TestTotal(t *testing.T) {
	s := NewStock()
	if got := s.Total(); got.String() != "0" {
		t.Fatalf("empty Total() = %s, want 0", got)
	}

	s.Add("apple", 3)
	s.Add("pear", 4)
	s.Add("apple", 2)
	if got := s.Total(); got.String() != "9" {
		t.Fatalf("Total() after additions = %s, want 9", got)
	}

	if err := s.Remove("pear", 6); err != nil {
		t.Fatalf("Remove(pear) = %v, want nil", err)
	}
	if got := s.Total(); got.String() != "3" {
		t.Fatalf("Total() after removal = %s, want 3", got)
	}
}

func TestTotalOutsideIntRange(t *testing.T) {
	max := int(^uint(0) >> 1)
	min := -max - 1
	tests := []struct {
		name       string
		quantities []int
		want       string
	}{
		{"above maximum", []int{max, 1}, new(big.Int).Add(big.NewInt(int64(max)), big.NewInt(1)).String()},
		{"below minimum", []int{min, -1}, new(big.Int).Sub(big.NewInt(int64(min)), big.NewInt(1)).String()},
		{"cancellation", []int{max, max, min}, fmt.Sprint(max - 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStock()
			for i, qty := range tt.quantities {
				s.Add(fmt.Sprint(i), qty)
			}
			if got := fmt.Sprint(s.Total()); got != tt.want {
				t.Fatalf("Total() = %s, want %s", got, tt.want)
			}
		})
	}
}
