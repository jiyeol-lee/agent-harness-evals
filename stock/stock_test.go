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

func TestListReturnsIndependentSnapshot(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	before := s.List()

	s.Add("apple", 2)
	after := s.List()
	if len(before) != 1 || len(after) != 1 || before[0].Count != 3 || after[0].Count != 5 {
		t.Fatalf("snapshots before and after Add = %v and %v, want [{apple 3}] and [{apple 5}]", before, after)
	}

	after[0] = ItemCount{Name: "pear", Count: 7}
	if got := s.List(); len(got) != 1 || got[0] != (ItemCount{Name: "apple", Count: 5}) {
		t.Fatalf("snapshot after changing returned item = %v, want [{apple 5}]", got)
	}
}

func TestConcurrentStockOperations(t *testing.T) {
	s := NewStock()
	s.Add("apple", 0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for qty := 1; qty <= 100; qty++ {
			s.Add("apple", qty)
			s.Count("apple")
		}
	}()

	for qty := 1; qty <= 100; qty++ {
		if err := s.Remove("apple", qty); err != nil {
			t.Error(err)
		}
		s.List()
	}
	<-done

	if got := s.Count("apple"); got != 0 {
		t.Errorf("Count(apple) after balanced adds and removes = %d, want 0", got)
	}
}
