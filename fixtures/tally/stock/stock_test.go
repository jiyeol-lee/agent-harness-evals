package stock

import (
	"path/filepath"
	"testing"
)

func TestAddAccumulates(t *testing.T) {
	s := New()
	s.Add("apple", 3)
	s.Add("apple", 2)

	if got := s.Count("apple"); got != 5 {
		t.Fatalf("Count(apple) = %d, want 5", got)
	}
}

func TestRemove(t *testing.T) {
	s := New()
	s.Add("apple", 10)

	if err := s.Remove("apple", 3); err != nil {
		t.Fatal(err)
	}
	if got := s.Count("apple"); got != 7 {
		t.Fatalf("Count(apple) = %d, want 7", got)
	}
}

func TestRemoveUnknown(t *testing.T) {
	s := New()

	if err := s.Remove("pear", 1); err != ErrUnknownItem {
		t.Fatalf("Remove(pear) = %v, want ErrUnknownItem", err)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stock.json")
	s := New()
	s.Add("apple", 4)
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if n := got.Count("apple"); n != 4 {
		t.Fatalf("Count(apple) after Load = %d, want 4", n)
	}
}
