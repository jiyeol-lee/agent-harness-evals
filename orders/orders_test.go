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

func TestQuoteAndPlaceUseItemOrder(t *testing.T) {
	items := []Item{
		{Name: "apple", UnitPrice: 0.05, Qty: 2},
		{Name: "pear", UnitPrice: 0.2, Qty: 1},
		{Name: "apple", UnitPrice: 0.15, Qty: 2},
	}
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 4)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 6 || pears != 3 {
			t.Errorf("stock at publication = (%d apples, %d pears), want (6, 3)", apples, pears)
		}
		published = append(published, e)
	})
	svc := New(st, bus)

	const want = 0.6000000000000001
	if got, err := svc.Quote(items); err != nil || got != want {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", got, err, want)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 10 || pears != 4 || len(published) != 0 {
		t.Fatalf("Quote changed stock or events: %d apples, %d pears, events = %v", apples, pears, published)
	}
	if got, err := svc.Place("o1", items); err != nil || got != want {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", got, err, want)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 4, "pear": 1}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Fatalf("events = %v, want %v", published, wantEvents)
	}
}

func TestPlaceStopsAtFirstReservationError(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 4)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) { t.Errorf("unexpected event: %v", e) })
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: 1, Qty: 1},
	})
	if got != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want (0, %v)", got, err, stock.ErrUnknownItem)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 4 {
		t.Fatalf("stock after error = (%d apples, %d pears), want (8, 4)", apples, pears)
	}
}
