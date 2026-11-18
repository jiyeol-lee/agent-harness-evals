// Package pricing computes order prices.
package pricing

// Line is one priced line of an order.
type Line struct {
	UnitPrice float64
	Qty       int
}

// Quote returns the total for lines after applying discountPct, from 0 to 100.
func Quote(lines []Line, discountPct float64) float64 {
	var total float64
	for _, l := range lines {
		total += l.UnitPrice * float64(l.Qty)
	}
	return total * (1 - discountPct/100)
}

// CalcPrice returns the price of qty units at unitPrice.
//
// Deprecated: use Quote, which prices a whole order and supports discounts.
func CalcPrice(unitPrice float64, qty int) float64 {
	return unitPrice * float64(qty)
}
