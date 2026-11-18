package pricing

import "testing"

func TestQuote(t *testing.T) {
	lines := []Line{{UnitPrice: 1.25, Qty: 2}, {UnitPrice: 2.5, Qty: 3}}
	for _, tt := range []struct {
		name     string
		lines    []Line
		discount float64
		want     float64
	}{
		{"no discount", lines, 0, 10},
		{"discount", lines, 25, 7.5},
		{"empty", nil, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := Quote(tt.lines, tt.discount); got != tt.want {
				t.Fatalf("quote = %v, want %v", got, tt.want)
			}
		})
	}
}
