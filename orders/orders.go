// Package orders prices orders and reserves their stock.
package orders

import (
	"errors"
	"fmt"

	"example.com/shop/events"
	"example.com/shop/pricing"
	"example.com/shop/stock"
)

// Item is one line of an order.
type Item struct {
	Name      string
	UnitPrice float64
	Qty       int
}

// Service prices and places orders.
type Service struct {
	stock *stock.Stock
	bus   *events.Bus
}

// New returns a Service that reserves from s and publishes to b.
func New(s *stock.Stock, b *events.Bus) *Service {
	return &Service{stock: s, bus: b}
}

// Quote returns the total price of items without reserving stock.
func (o *Service) Quote(items []Item) (float64, error) {
	if len(items) == 0 {
		return 0, errors.New("order has no items")
	}
	lines := make([]pricing.Line, len(items))
	for i, it := range items {
		if it.Qty <= 0 {
			return 0, fmt.Errorf("item %s: quantity must be positive", it.Name)
		}
		if it.UnitPrice < 0 {
			return 0, fmt.Errorf("item %s: price must not be negative", it.Name)
		}
		lines[i] = pricing.Line{UnitPrice: it.UnitPrice, Qty: it.Qty}
	}

	return pricing.Quote(lines, 0), nil
}

// Place reserves stock for items, publishes OrderPlaced, and returns the total.
func (o *Service) Place(id string, items []Item) (float64, error) {
	if len(items) == 0 {
		return 0, errors.New("order has no items")
	}
	for _, it := range items {
		if it.Qty <= 0 {
			return 0, fmt.Errorf("item %s: quantity must be positive", it.Name)
		}
		if it.UnitPrice < 0 {
			return 0, fmt.Errorf("item %s: price must not be negative", it.Name)
		}
	}

	lines := make([]pricing.Line, len(items))
	reserved := map[string]int{}
	for i, it := range items {
		if err := o.stock.Remove(it.Name, it.Qty); err != nil {
			return 0, fmt.Errorf("reserve %s: %w", it.Name, err)
		}
		reserved[it.Name] += it.Qty
		lines[i] = pricing.Line{UnitPrice: it.UnitPrice, Qty: it.Qty}
	}

	total := pricing.Quote(lines, 0)
	o.bus.Publish(events.OrderPlaced{OrderID: id, Items: reserved})
	return total, nil
}
