// Package pricing computes order prices.
package pricing

// CalcPrice returns the price of qty units at unitPrice.
func CalcPrice(unitPrice float64, qty int) float64 {
	return unitPrice * float64(qty)
}
