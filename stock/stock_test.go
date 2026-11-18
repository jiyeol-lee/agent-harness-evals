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

	s.Add("pear", 0)
	s.Add("apple", 3)
	s.Add("banana", -2)
	s.Add("apple", 2)
	if err := s.Remove("apple", 1); err != nil {
		t.Fatal(err)
	}
	want := []Item{{Name: "apple", Count: 4}, {Name: "banana", Count: -2}, {Name: "pear", Count: 0}}
	got := s.List()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %#v, want %#v", got, want)
	}
	got[0].Count = 99
	s.Add("apple", 1)
	if got := s.List(); got[0].Count != 5 {
		t.Fatalf("List() after snapshot mutation and update = %#v, want apple count 5", got)
	}
	if got[0].Count != 99 {
		t.Fatalf("previous snapshot changed: %#v", got)
	}
}

func TestConcurrentStock(t *testing.T) {
	s := NewStock()
	s.Add("apple", 0)
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
}
