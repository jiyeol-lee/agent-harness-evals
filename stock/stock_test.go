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

func TestListReturnsDetachedSnapshot(t *testing.T) {
	s := NewStock()
	if got := s.List(); got == nil || len(got) != 0 {
		t.Fatalf("empty List() = %#v, want non-nil empty slice", got)
	}

	s.Add("apple", 3)
	snapshot := s.List()
	s.Add("apple", 2)
	if got := snapshot[0]; got != (Item{Name: "apple", Count: 3}) {
		t.Fatalf("snapshot after Add = %+v, want apple with count 3", got)
	}
	snapshot[0] = Item{Name: "pear", Count: 1}
	if got := s.List()[0]; got != (Item{Name: "apple", Count: 5}) {
		t.Fatalf("List() after snapshot edit = %+v, want apple with count 5", got)
	}
}

func TestListWithConcurrentStockOperations(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1)

	var writes sync.WaitGroup
	writes.Go(func() {
		for i := 1; i <= 100; i++ {
			s.Add("apple", i)
			if err := s.Remove("apple", i); err != nil {
				t.Error(err)
			}
		}
	})
	for range 100 {
		s.List()
		s.Count("apple")
	}
	writes.Wait()

	if got := s.List()[0]; got != (Item{Name: "apple", Count: 1}) {
		t.Fatalf("List() after concurrent operations = %+v, want apple with count 1", got)
	}
}
