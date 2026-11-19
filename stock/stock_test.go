package stock

import (
	"sync"
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

func TestTotalEmpty(t *testing.T) {
	s := NewStock()

	if got := s.Total(); got != 0 {
		t.Fatalf("Total() = %d, want 0", got)
	}
}

func TestTotalAfterAddAndRemove(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 4)
	s.Add("apple", 2)

	if got := s.Total(); got != 9 {
		t.Fatalf("Total() after Add = %d, want 9", got)
	}

	if err := s.Remove("apple", 6); err != nil {
		t.Fatalf("Remove(apple, 6) = %v, want nil", err)
	}
	if got := s.Total(); got != 3 {
		t.Fatalf("Total() after Remove = %d, want 3", got)
	}
}

func TestStockConcurrentReadsAndWrites(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 4)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, name := range []string{"apple", "pear"} {
		wg.Go(func() {
			<-start
			for i := range 100 {
				qty := 1 + i%2
				s.Add(name, qty)
				if err := s.Remove(name, qty); err != nil {
					t.Errorf("Remove(%s, %d) = %v, want nil", name, qty, err)
					return
				}
			}
		})
	}
	wg.Go(func() {
		<-start
		for range 100 {
			s.Count("apple")
			s.Total()
		}
	})
	close(start)
	wg.Wait()

	if got := s.Total(); got != 7 {
		t.Fatalf("Total() after concurrent updates = %d, want 7", got)
	}
}
