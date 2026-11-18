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

func TestTotalEmpty(t *testing.T) {
	var zero Stock
	if got := zero.Total(); got != 0 {
		t.Fatalf("zero Stock.Total() = %d, want 0", got)
	}
	if got := NewStock().Total(); got != 0 {
		t.Fatalf("NewStock().Total() = %d, want 0", got)
	}
}

func TestTotalCurrentQuantities(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	s.Add("pear", 5)

	if got := s.Total(); got != 8 {
		t.Fatalf("Total() = %d, want 8", got)
	}

	s.Add("apple", 2)
	if err := s.Remove("pear", 7); err != nil {
		t.Fatalf("Remove(pear) = %v, want nil", err)
	}
	if got := s.Total(); got != 3 {
		t.Fatalf("Total() after updates = %d, want 3", got)
	}
}

func TestTotalConcurrentAdd(t *testing.T) {
	s := NewStock()
	s.Add("apple", 1)

	const additions = 100
	done := make(chan struct{})
	go func() {
		for i := 0; i < additions; i++ {
			s.Add("apple", 1)
		}
		close(done)
	}()

	for i := 0; i < additions; i++ {
		s.Total()
	}
	<-done

	if got, want := s.Total(), 1+additions; got != want {
		t.Fatalf("Total() after concurrent adds = %d, want %d", got, want)
	}
}
