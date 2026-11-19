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
		t.Fatalf("Total() = %d, want 0", got)
	}
}

func TestTotalAfterChanges(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 5)

	if got := s.Total(); got != 8 {
		t.Fatalf("Total() = %d, want 8", got)
	}

	s.Add("apple", 2)
	if err := s.Remove("pear", 1); err != nil {
		t.Fatal(err)
	}

	if got := s.Total(); got != 9 {
		t.Fatalf("Total() after changes = %d, want 9", got)
	}
}
