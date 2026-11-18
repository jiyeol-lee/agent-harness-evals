// Package events is an in-process publish and subscribe bus.
package events

// OrderPlaced is published after an order reserves its stock.
type OrderPlaced struct {
	OrderID string
	Items   map[string]int
}

// Bus delivers events to subscribers synchronously, in subscription order.
type Bus struct {
	subs []func(OrderPlaced)
}

// Subscribe registers fn to receive every OrderPlaced event.
func (b *Bus) Subscribe(fn func(OrderPlaced)) {
	b.subs = append(b.subs, fn)
}

// Publish delivers e to every subscriber.
func (b *Bus) Publish(e OrderPlaced) {
	for _, fn := range b.subs {
		fn(e)
	}
}
