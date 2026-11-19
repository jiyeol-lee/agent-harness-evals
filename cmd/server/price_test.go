package main

import (
	"math"
	"testing"
)

func TestPricePreservesSignedZero(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	for _, tc := range []struct {
		name string
		unit float64
		qty  int
		want float64
	}{
		{"negative zero price", negativeZero, 1, negativeZero},
		{"negative quantity", 0, -3, negativeZero},
		{"zero quantity", -2.5, 0, negativeZero},
		{"two negatives", negativeZero, -1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := quotePrice(tc.unit, tc.qty); math.Float64bits(got) != math.Float64bits(tc.want) {
				t.Errorf("price = %v, want %v", got, tc.want)
			}
		})
	}
}
