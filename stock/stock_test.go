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

func TestList(t *testing.T) {
	s := NewStock()
	if got := s.List(); got == nil || len(got) != 0 {
		t.Fatalf("empty List() = %#v, want non-nil empty slice", got)
	}
	s.Add("pear", -2)
	s.Add("banana", 0)
	s.Add("apple", 3)
	want := []Item{{Name: "apple", Count: 3}, {Name: "banana", Count: 0}, {Name: "pear", Count: -2}}
	got := s.List()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
	got[0].Count = 100
	s.Add("pear", 1)
	if got[2].Count != -2 {
		t.Fatalf("snapshot changed after Add: %#v", got)
	}
	want[2].Count = -1
	if got := s.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() after mutation = %#v, want %#v", got, want)
	}
}

func TestConcurrentStockAccess(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1000)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Add("apple", 1)
				_ = s.Remove("apple", 1)
				_ = s.Count("apple")
				_ = s.List()
			}
		}()
	}
	wg.Wait()
	if got := s.List(); len(got) != 1 || got[0].Name != "apple" {
		t.Fatalf("List() = %#v, want one apple", got)
	}
}
