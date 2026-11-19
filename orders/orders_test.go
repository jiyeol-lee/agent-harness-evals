package orders

import (
	"errors"
	"math"
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

func TestQuoteAndPlaceKeepItemOrder(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	published := 0
	bus.Subscribe(func(e events.OrderPlaced) {
		published++
		if e.OrderID != "o1" || len(e.Items) != 1 || e.Items["apple"] != 4 {
			t.Errorf("unexpected event: %+v", e)
		}
		if left := st.Count("apple"); left != 6 {
			t.Errorf("apple left at publication = %d, want 6", left)
		}
	})
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 1, Qty: 1},
		{Name: "apple", UnitPrice: 5e15, Qty: 2},
		{Name: "apple", UnitPrice: 1, Qty: 1},
	}

	const want = 1e16
	if got, err := svc.Quote(items); err != nil || got != want {
		t.Fatalf("Quote() = %v, %v; want %v, nil", got, err, want)
	}
	if left := st.Count("apple"); left != 10 || published != 0 {
		t.Fatalf("after Quote: apple left = %d, events = %d; want 10, 0", left, published)
	}
	if got, err := svc.Place("o1", items); err != nil || got != want {
		t.Fatalf("Place() = %v, %v; want %v, nil", got, err, want)
	}
	if published != 1 {
		t.Fatalf("events = %d, want 1", published)
	}
}

func TestQuoteAndPlaceReturnPositiveZero(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	svc := New(st, &events.Bus{})
	items := []Item{{Name: "apple", UnitPrice: math.Copysign(0, -1), Qty: 1}}

	if got, err := svc.Quote(items); err != nil || got != 0 || math.Signbit(got) {
		t.Errorf("Quote() = %v, %v; want +0, nil", got, err)
	}
	if got, err := svc.Place("o1", items); err != nil || got != 0 || math.Signbit(got) {
		t.Errorf("Place() = %v, %v; want +0, nil", got, err)
	}
}

func TestPlaceFailureDoesNotPublish(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("unexpected event: %+v", e)
	})
	svc := New(st, bus)
	apple := Item{Name: "apple", UnitPrice: 2, Qty: 1}

	invalid := []Item{apple, {Name: "missing", Qty: 0}}
	if _, err := svc.Place("invalid", invalid); err == nil || errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place(invalid) error = %v, want validation error", err)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("after validation failure: apple left = %d, want 10", left)
	}

	missing := []Item{apple, {Name: "missing", Qty: 1}}
	if _, err := svc.Place("missing", missing); !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place(missing) error = %v, want %v", err, stock.ErrUnknownItem)
	}
	if left := st.Count("apple"); left != 9 {
		t.Fatalf("after reservation failure: apple left = %d, want 9", left)
	}
}
