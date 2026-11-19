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

func TestSnapshotConcurrent(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1)

	var writer sync.WaitGroup
	writer.Add(1)
	go func() {
		defer writer.Done()
		for qty := 1; qty <= 100; qty++ {
			s.Add("apple", qty)
			if err := s.Remove("apple", qty); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	defer writer.Wait()

	for i := 0; i < 100; i++ {
		if got := s.Snapshot(); len(got) != 1 || got[0].Name != "apple" || got[0].Count < 1 {
			t.Fatalf("Snapshot() = %v, want an apple count of at least 1", got)
		}
		if got := s.Count("apple"); got < 1 {
			t.Fatalf("Count(apple) = %d, want at least 1", got)
		}
	}
}
