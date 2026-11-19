package orders

import (
	"errors"
	"testing"

	"example.com/shop/events"
	"example.com/shop/pricing"
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

func TestQuotePreservesInputOrder(t *testing.T) {
	items := []Item{
		{Name: "a", UnitPrice: 0.1, Qty: 1},
		{Name: "b", UnitPrice: 0.2, Qty: 1},
		{Name: "c", UnitPrice: 0.3, Qty: 1},
	}
	svc := New(stock.NewStock(), &events.Bus{})

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.6000000000000001; got != want {
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
	if want := pricing.Quote([]pricing.Line{{UnitPrice: 2.5, Qty: 2}}, 0); got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}

func TestPlacePublishesAfterReservations(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	published := 0
	bus.Subscribe(func(e events.OrderPlaced) {
		published++
		if e.OrderID != "o1" || len(e.Items) != 2 || e.Items["apple"] != 2 || e.Items["pear"] != 3 {
			t.Errorf("unexpected event: %+v", e)
		}
		if left := st.Count("apple"); left != 8 {
			t.Errorf("apple left at publication = %d, want 8", left)
		}
		if left := st.Count("pear"); left != 7 {
			t.Errorf("pear left at publication = %d, want 7", left)
		}
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1, Qty: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 8 {
		t.Fatalf("total = %v, want 8", got)
	}
	if published != 1 {
		t.Fatalf("published %d events, want 1", published)
	}
}

func TestPlaceFailure(t *testing.T) {
	tests := []struct {
		name        string
		second      Item
		wantUnknown bool
		wantLeft    int
	}{
		{"invalid quantity prevents reservation", Item{Name: "missing", UnitPrice: 2, Qty: 0}, false, 10},
		{"negative price prevents reservation", Item{Name: "missing", UnitPrice: -2, Qty: 1}, false, 10},
		{"unknown item leaves prior reservation", Item{Name: "missing", UnitPrice: 2, Qty: 1}, true, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			bus := &events.Bus{}
			bus.Subscribe(func(e events.OrderPlaced) {
				t.Errorf("published on failure: %+v", e)
			})
			svc := New(st, bus)

			got, err := svc.Place("o1", []Item{{Name: "apple", UnitPrice: 2.5, Qty: 2}, tt.second})
			if got != 0 || err == nil || errors.Is(err, stock.ErrUnknownItem) != tt.wantUnknown {
				t.Fatalf("Place = (%v, %v), want zero total and error (unknown item: %t)", got, err, tt.wantUnknown)
			}
			if left := st.Count("apple"); left != tt.wantLeft {
				t.Fatalf("apple left = %d, want %d", left, tt.wantLeft)
			}
		})
	}
}
