package stock

import (
	"reflect"
	"testing"
)

func TestList(t *testing.T) {
	s := NewStock()
	if got := s.List(); got == nil || len(got) != 0 {
		t.Fatalf("empty List() = %#v, want empty slice", got)
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
