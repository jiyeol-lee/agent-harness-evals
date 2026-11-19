package orders

import (
	"errors"
	"reflect"
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

func TestQuoteAndPlacePreserveLineOrder(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) { published = append(published, e) })
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 1e16, Qty: 1},
		{Name: "pear", UnitPrice: 1, Qty: 1},
		{Name: "apple", UnitPrice: 0.5, Qty: 2},
	}
	// Adding the two small line totals separately rounds each away.
	const want = 1e16
	if got, err := svc.Quote(items); err != nil || got != want {
		t.Fatalf("Quote = %v, %v; want %v, nil", got, err, want)
	}
	if apple, pear := st.Count("apple"), st.Count("pear"); apple != 10 || pear != 10 || len(published) != 0 {
		t.Fatalf("after Quote: apple = %d, pear = %d, events = %v", apple, pear, published)
	}
	if got, err := svc.Place("o1", items); err != nil || got != want {
		t.Fatalf("Place = %v, %v; want %v, nil", got, err, want)
	}
	if apple, pear := st.Count("apple"), st.Count("pear"); apple != 7 || pear != 9 {
		t.Fatalf("after Place: apple = %d, pear = %d; want 7, 9", apple, pear)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 1}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Fatalf("events = %v, want %v", published, wantEvents)
	}
}

func TestPlaceValidatesBeforeReservingAndStopsOnMissingStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) { t.Errorf("unexpected event: %v", e) })
	svc := New(st, bus)
	items := []Item{{Name: "apple", UnitPrice: 2, Qty: 2}, {Name: "missing", UnitPrice: -1, Qty: 1}}

	if got, err := svc.Place("o1", items); err == nil || got != 0 {
		t.Fatalf("Place with invalid price = %v, %v; want zero total and an error", got, err)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left after validation failure = %d, want 10", left)
	}

	items[1].UnitPrice = 1
	if got, err := svc.Place("o2", items); !errors.Is(err, stock.ErrUnknownItem) || got != 0 {
		t.Fatalf("Place with unknown stock = %v, %v; want zero total and unknown item", got, err)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left after reservation failure = %d, want 8", left)
	}
}
