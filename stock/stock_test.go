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
		t.Fatalf("new stock Total() = %d, want 0", got)
	}

	var zero Stock
	if got := zero.Total(); got != 0 {
		t.Fatalf("zero-value stock Total() = %d, want 0", got)
	}
}

func TestTotalAfterAddsAndRemoval(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("apple", 2)
	s.Add("pear", 4)
	s.Add("plum", -2)

	if got := s.Total(); got != 7 {
		t.Fatalf("Total() after adds = %d, want 7", got)
	}

	if err := s.Remove("pear", 8); err != nil {
		t.Fatalf("Remove(pear, 8): %v", err)
	}
	if got := s.Total(); got != -1 {
		t.Fatalf("Total() after removal = %d, want -1", got)
	}
}

func TestTotalConcurrentWithAddAndRemove(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1)
	const iterations = 1000

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < iterations; i++ {
			s.Total()
		}
	}()

	for qty := 1; qty <= iterations; qty++ {
		s.Add("apple", qty)
		if err := s.Remove("apple", qty); err != nil {
			t.Errorf("Remove(apple, %d): %v", qty, err)
			break
		}
	}
	<-done

	if got := s.Total(); got != 1 {
		t.Fatalf("Total() after balanced adds and removals = %d, want 1", got)
	}
}
