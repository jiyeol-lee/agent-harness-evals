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

func TestTotal(t *testing.T) {
	s := NewStock()
	if got := s.Total(); got != 0 {
		t.Fatalf("empty Total() = %d, want 0", got)
	}

	s.Add("apple", 3)
	s.Add("pear", 7)
	s.Add("apple", 2)
	if got := s.Total(); got != 12 {
		t.Fatalf("Total() after additions = %d, want 12", got)
	}

	if err := s.Remove("pear", 13); err != nil {
		t.Fatalf("Remove(pear) = %v, want nil", err)
	}
	if got := s.Total(); got != -1 {
		t.Fatalf("Total() after removal = %d, want -1", got)
	}
}

func TestStockReadsWhileUpdating(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1)
	s.Add("pear", 1)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 100 {
			for _, name := range []string{"apple", "pear"} {
				s.Add(name, 1)
				if err := s.Remove(name, 1); err != nil {
					t.Errorf("Remove(%s) = %v, want nil", name, err)
					return
				}
			}
		}
	}()

	for range 100 {
		if got := s.Total(); got != 2 && got != 3 {
			t.Errorf("Total() during updates = %d, want 2 or 3", got)
		}
		if got := s.Count("apple"); got != 1 && got != 2 {
			t.Errorf("Count(apple) during updates = %d, want 1 or 2", got)
		}
	}
	<-done
}
