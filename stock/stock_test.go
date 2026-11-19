package stock

import (
	"reflect"
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
		t.Fatalf("empty List() = %#v, want []Item{}", got)
	}

	s.Add("pear", 1)
	s.Add("apple", 2)
	s.Add("apple", 3)
	if err := s.Remove("pear", 1); err != nil {
		t.Fatal(err)
	}

	want := []Item{{Name: "apple", Count: 5}, {Name: "pear", Count: 0}}
	if got := s.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
}

func TestListReturnsIndependentSnapshot(t *testing.T) {
	s := NewStock()
	s.Add("apple", 2)
	first := s.List()
	s.Add("apple", 3)
	second := s.List()

	if got := first[0].Count; got != 2 {
		t.Fatalf("first snapshot count = %d after Add, want 2", got)
	}
	first[0] = Item{Name: "pear", Count: 9}
	want := Item{Name: "apple", Count: 5}
	if got := second[0]; got != want {
		t.Errorf("second snapshot item = %v after editing first, want %v", got, want)
	}
	if got := s.List()[0]; got != want {
		t.Errorf("List()[0] = %v after editing first snapshot, want %v", got, want)
	}
}

func TestConcurrentStockAccess(t *testing.T) {
	const additions = 100
	s := NewStock()
	s.Add("apple", 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < additions; i++ {
			s.Add("apple", 1)
		}
	}()

	err := s.Remove("apple", 1)
	for i := 0; i < additions; i++ {
		s.Count("apple")
		s.List()
	}
	<-done
	if err != nil {
		t.Fatal(err)
	}
	want := []Item{{Name: "apple", Count: additions}}
	if got := s.List(); !reflect.DeepEqual(got, want) {
		t.Fatalf("List() = %v, want %v", got, want)
	}
}
