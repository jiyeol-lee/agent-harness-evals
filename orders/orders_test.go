package orders

import (
	"errors"
	"reflect"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestQuoteAndPlaceOrderedItems(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 5)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 7 || pears != 4 {
			t.Errorf("stock at publish = (%d, %d), want (7, 4)", apples, pears)
		}
		published = append(published, e)
	})
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 0.05, Qty: 2},
		{Name: "apple", UnitPrice: 0.2, Qty: 1},
		{Name: "pear", UnitPrice: 0.3, Qty: 1},
	}
	const wantTotal = 0.6000000000000001

	if total, err := svc.Quote(items); err != nil || total != wantTotal {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", total, err, wantTotal)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 10 || pears != 5 || len(published) != 0 {
		t.Fatalf("stock after Quote = (%d, %d), events = %v; want (10, 5) and no events", apples, pears, published)
	}
	if total, err := svc.Place("o1", items); err != nil || total != wantTotal {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", total, err, wantTotal)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 1}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Errorf("events = %v, want %v", published, wantEvents)
	}
}

func TestPlaceValidationBeforeReservation(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) { t.Errorf("unexpected event: %v", e) })
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: -1, Qty: 0},
	}

	if total, err := svc.Place("o1", items); total != 0 || err == nil || err.Error() != "item pear: quantity must be positive" {
		t.Fatalf("Place with invalid item = (%v, %v), want (0, quantity error for pear)", total, err)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left after validation error = %d, want 10", left)
	}
	if total, err := svc.Place("o1", items[:2]); total != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place with unknown item = (%v, %v), want (0, ErrUnknownItem)", total, err)
	}
	if left := st.Count("apple"); left != 8 {
		t.Errorf("apple left after reservation error = %d, want 8", left)
	}
}

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
