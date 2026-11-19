package orders

import (
	"errors"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestOrderTotal(t *testing.T) {
	prices := map[string]float64{"a": 0.1, "b": 0.2, "c": 0.3}
	var items []Item
	for name, price := range prices {
		items = append(items, Item{Name: name, UnitPrice: price, Qty: 1})
	}
	svc := New(stock.NewStock(), &events.Bus{})

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.1 + 0.2 + 0.3; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	svc := New(st, &events.Bus{})

	got, err := svc.Place("o1", []Item{{Name: "apple", UnitPrice: 2.5, Qty: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if want := 5.0; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}

func TestOrderPricingPreservesItemOrder(t *testing.T) {
	items := []Item{
		{Name: "a", UnitPrice: 1e16, Qty: 1},
		{Name: "b", UnitPrice: 0.5, Qty: 2},
		{Name: "c", UnitPrice: 1, Qty: 1},
	}
	st := stock.NewStock()
	for _, item := range items {
		st.Add(item.Name, 10)
	}
	bus := &events.Bus{}
	var published int
	bus.Subscribe(func(e events.OrderPlaced) {
		published++
		if e.OrderID != "o1" || len(e.Items) != 3 {
			t.Errorf("unexpected event: %+v", e)
		}
		for _, item := range items {
			if e.Items[item.Name] != item.Qty {
				t.Errorf("event quantity for %s = %d, want %d", item.Name, e.Items[item.Name], item.Qty)
			}
			if left := st.Count(item.Name); left != 10-item.Qty {
				t.Errorf("%s left at publish = %d, want %d", item.Name, left, 10-item.Qty)
			}
		}
	})
	svc := New(st, bus)

	// Each trailing 1 rounds away when added separately to 1e16.
	const want = 1e16
	if got, err := svc.Quote(items); got != want || err != nil {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", got, err, want)
	}
	for _, item := range items {
		if left := st.Count(item.Name); left != 10 {
			t.Errorf("%s left after Quote = %d, want 10", item.Name, left)
		}
	}
	if published != 0 {
		t.Fatalf("%d events published by Quote, want 0", published)
	}
	if got, err := svc.Place("o1", items); got != want || err != nil {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", got, err, want)
	}
	if published != 1 {
		t.Fatalf("%d events published by Place, want 1", published)
	}
}

func TestPlaceReservationFailureKeepsEarlierReservation(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("published after reservation failure: %+v", e)
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 1, Qty: 1},
	})
	if got != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want (0, wrapped ErrUnknownItem)", got, err)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}
