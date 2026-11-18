package pricing

import "testing"

func TestQuoteAppliesDiscountToOrderTotal(t *testing.T) {
	lines := []Line{{UnitPrice: 2, Qty: 3}, {UnitPrice: 4, Qty: 1}}

	if got := Quote(lines, 25); got != 7.5 {
		t.Fatalf("Quote = %v, want 7.5", got)
	}
}
