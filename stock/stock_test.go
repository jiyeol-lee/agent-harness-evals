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

func TestTotalEmpty(t *testing.T) {
	if got := NewStock().Total(); got != 0 {
		t.Fatalf("NewStock().Total() = %d, want 0", got)
	}

	var s Stock
	if got := s.Total(); got != 0 {
		t.Fatalf("zero-value Stock.Total() = %d, want 0", got)
	}
}

func TestTotalTracksCurrentQuantities(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 2)
	if got := s.Total(); got != 5 {
		t.Fatalf("Total() with two items = %d, want 5", got)
	}

	s.Add("apple", 4)
	if got := s.Total(); got != 9 {
		t.Fatalf("Total() after adding to an existing item = %d, want 9", got)
	}

	if err := s.Remove("apple", 1); err != nil {
		t.Fatalf("Remove(apple) = %v, want nil", err)
	}
	if got := s.Total(); got != 8 {
		t.Fatalf("Total() after removal = %d, want 8", got)
	}

	s.Add("plum", -10)
	if got := s.Total(); got != -2 {
		t.Fatalf("Total() with a negative quantity = %d, want -2", got)
	}
}
