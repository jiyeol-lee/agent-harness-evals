package stock

import "testing"

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

func TestTotalZeroValue(t *testing.T) {
	var s Stock

	if got := s.Total(); got != 0 {
		t.Fatalf("Total() = %d for zero-value Stock, want 0", got)
	}
}

func TestTotalAfterStockChanges(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 2)
	if got := s.Total(); got != 5 {
		t.Fatalf("Total() = %d after adding two items, want 5", got)
	}

	s.Add("apple", 4)
	if got := s.Total(); got != 9 {
		t.Fatalf("Total() = %d after adding more apples, want 9", got)
	}

	if err := s.Remove("pear", 12); err != nil {
		t.Fatalf("Remove(pear, 12) = %v, want nil", err)
	}
	if got := s.Total(); got != -3 {
		t.Fatalf("Total() = %d after removing pears, want -3", got)
	}
}
