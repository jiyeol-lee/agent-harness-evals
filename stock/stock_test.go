package stock

import (
	"reflect"
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

func TestSnapshot(t *testing.T) {
	s := NewStock()
	if got := s.Snapshot(); got == nil || len(got) != 0 {
		t.Fatalf("empty Snapshot() = %#v, want non-nil empty slice", got)
	}
	s.Add("pear", 2)
	s.Add("apple", 0)
	first := s.Snapshot()
	want := []ItemCount{{Name: "apple", Count: 0}, {Name: "pear", Count: 2}}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("Snapshot() = %#v, want %#v", first, want)
	}
	first[0].Count = 99
	s.Add("pear", 1)
	if first[1].Count != 2 {
		t.Fatalf("earlier snapshot changed: %#v", first)
	}
	want[1].Count = 3
	if got := s.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot() = %#v, want %#v", got, want)
	}
}

func TestConcurrentStockAccess(t *testing.T) {
	s := NewStock()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Add("apple", 1)
				_ = s.Remove("apple", 1)
				_ = s.Count("apple")
				_ = s.Snapshot()
			}
		}()
	}
	wg.Wait()
}
