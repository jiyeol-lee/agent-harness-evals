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

func TestRemoveInsufficientStock(t *testing.T) {
	s := NewStock()
	s.Add("apple", 2)

	for attempt := 1; attempt <= 2; attempt++ {
		err := s.Remove("apple", 3)
		if got := s.Count("apple"); got != 2 {
			t.Fatalf("after attempt %d, Count(apple) = %d, want 2", attempt, got)
		}
		if err != ErrInsufficientStock {
			t.Fatalf("attempt %d: Remove(apple, 3) = %v, want ErrInsufficientStock", attempt, err)
		}
	}
}

func TestRemoveExactStock(t *testing.T) {
	s := NewStock()
	s.Add("apple", 2)

	if err := s.Remove("apple", 2); err != nil {
		t.Fatalf("Remove(apple, 2) = %v, want nil", err)
	}
	if got := s.Count("apple"); got != 0 {
		t.Fatalf("Count(apple) = %d, want 0", got)
	}
	if err := s.Remove("apple", 2); err != ErrInsufficientStock {
		t.Fatalf("Remove(apple, 2) with no stock left = %v, want ErrInsufficientStock", err)
	}
}
