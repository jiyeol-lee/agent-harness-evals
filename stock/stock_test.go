package stock

import (
	"fmt"
	"math"
	"strconv"
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

func TestTotal(t *testing.T) {
	s := NewStock()
	if got := s.Total().String(); got != "0" {
		t.Fatalf("Total() for empty stock = %s, want 0", got)
	}

	s.Add("apple", 3)
	s.Add("pear", 4)
	if got := s.Total().String(); got != "7" {
		t.Fatalf("Total() after adding items = %s, want 7", got)
	}

	s.Add("apple", 2)
	if err := s.Remove("pear", 6); err != nil {
		t.Fatalf("Remove(pear, 6) = %v, want nil", err)
	}
	if got := s.Total().String(); got != "3" {
		t.Fatalf("Total() after changing quantities = %s, want 3", got)
	}
}

func TestTotalBeyondIntBounds(t *testing.T) {
	aboveMax, belowMin := "18446744073709551614", "-18446744073709551464"
	if strconv.IntSize == 32 {
		aboveMax, belowMin = "4294967294", "-4294967144"
	}

	for _, tt := range []struct {
		name          string
		apples, pears int
		want          string
	}{
		{"above max", math.MaxInt, math.MaxInt, aboveMax},
		{"below min", math.MinInt + 101, math.MinInt + 51, belowMin},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStock()
			s.Add("apple", tt.apples)
			s.Add("pear", tt.pears)
			if got := fmt.Sprint(s.Total()); got != tt.want {
				t.Fatalf("Total() = %s, want %s", got, tt.want)
			}
		})
	}
}
