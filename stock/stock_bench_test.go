package stock

import (
	"fmt"
	"testing"
)

func BenchmarkCount(b *testing.B) {
	s := NewStock()
	for i := range 100_000 {
		s.Add(fmt.Sprintf("item-%d", i), 1)
	}

	b.ResetTimer()
	for range b.N {
		s.Count("item-99999")
	}
}
