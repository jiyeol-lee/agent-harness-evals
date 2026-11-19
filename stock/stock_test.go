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

func TestTotalAfterUpdates(t *testing.T) {
	s := NewStock()
	s.Add("apple", 5)
	s.Add("pear", -2)

	if got := s.Total(); got != 3 {
		t.Fatalf("Total() after adding two items = %d, want 3", got)
	}

	s.Add("apple", 2)
	if got := s.Total(); got != 5 {
		t.Fatalf("Total() after updating apple = %d, want 5", got)
	}

	if err := s.Remove("apple", 6); err != nil {
		t.Fatalf("Remove(apple) = %v, want nil", err)
	}
	if got := s.Total(); got != -1 {
		t.Fatalf("Total() after removing apples = %d, want -1", got)
	}
}

func TestTotalDuringRemovals(t *testing.T) {
	s := NewStock()
	s.Add("apple", 150)

	start := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		<-start
		for i := 0; i < 100; i++ {
			if got := s.Total(); got < 0 || got > 150 {
				t.Errorf("Total() = %d, want between 0 and 150", got)
				return
			}
		}
	}()

	close(start)
	for i := 0; i < 50; i++ {
		// Alternate quantities so removal deduplication does not skip a call.
		for _, qty := range []int{1, 2} {
			if err := s.Remove("apple", qty); err != nil {
				t.Errorf("Remove(apple, %d) = %v, want nil", qty, err)
			}
		}
	}
	<-done

	if got := s.Total(); got != 0 {
		t.Fatalf("Total() after all removals = %d, want 0", got)
	}
}
