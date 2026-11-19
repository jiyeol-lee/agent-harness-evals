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

func TestSnapshotIndependent(t *testing.T) {
	s := NewStock()
	s.Add("apple", 3)
	first := s.Snapshot()
	s.Add("apple", 2)
	second := s.Snapshot()

	if len(first) != 1 || first[0] != (ItemCount{Name: "apple", Count: 3}) {
		t.Fatalf("first Snapshot() = %v, want [{apple 3}]", first)
	}
	if len(second) != 1 || second[0] != (ItemCount{Name: "apple", Count: 5}) {
		t.Fatalf("second Snapshot() = %v, want [{apple 5}]", second)
	}
	first[0].Count = -1
	if got := s.Count("apple"); got != 5 {
		t.Errorf("Count(apple) = %d after modifying snapshot, want 5", got)
	}
	if second[0].Count != 5 {
		t.Errorf("second snapshot count = %d after modifying first, want 5", second[0].Count)
	}
}

func TestSnapshotWhileUpdating(t *testing.T) {
	s := NewStock()
	s.Add("apple", 0)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 100 {
			s.Add("apple", 1)
		}
	}()
	for range 100 {
		s.Snapshot()
	}
	<-done

	got := s.Snapshot()
	if len(got) != 1 || got[0] != (ItemCount{Name: "apple", Count: 100}) {
		t.Errorf("Snapshot() after updates = %v, want [{apple 100}]", got)
	}
}
