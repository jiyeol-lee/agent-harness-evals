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
	s := NewStock()

	if got := s.Total(); got != 0 {
		t.Fatalf("Total() on empty stock = %d, want 0", got)
	}
}

func TestTotalAfterAddAndRemove(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("apple", 2)
	s.Add("pear", 4)

	if got := s.Total(); got != 9 {
		t.Fatalf("Total() after additions = %d, want 9", got)
	}

	if err := s.Remove("apple", 10); err != nil {
		t.Fatalf("Remove(apple, 10) = %v, want nil", err)
	}
	if got := s.Total(); got != -1 {
		t.Fatalf("Total() after removal = %d, want -1", got)
	}
}
