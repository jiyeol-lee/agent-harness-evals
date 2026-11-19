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

func TestQuotePricesWithoutReserving(t *testing.T) {
	items := []Item{
		{Name: "a", UnitPrice: 0.1, Qty: 1},
		{Name: "b", UnitPrice: 0.2, Qty: 1},
		{Name: "c", UnitPrice: 0.3, Qty: 1},
	}
	st := stock.NewStock()
	st.Add("a", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("Quote published %+v", e)
	})
	svc := New(st, bus)

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.6000000000000001; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if left := st.Count("a"); left != 10 {
		t.Fatalf("a left = %d, want 10", left)
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 7 || pears != 6 {
			t.Errorf("stock at publication = (%d apples, %d pears), want (7, 6)", apples, pears)
		}
		published = append(published, e)
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 4},
		{Name: "apple", UnitPrice: 2.5, Qty: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := 12.5; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	want := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 4}}}
	if !reflect.DeepEqual(published, want) {
		t.Fatalf("published = %+v, want %+v", published, want)
	}
}

func TestPlaceValidatesBeforeReserving(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("Place published %+v for an invalid order", e)
	})
	svc := New(st, bus)

	total, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 0},
	})
	if total != 0 || err == nil || errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want zero and a validation error", total, err)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left = %d, want 10", left)
	}
}

func TestPlaceStopsAtReservationError(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("Place published %+v after a reservation error", e)
	})
	svc := New(st, bus)

	total, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: 1.25, Qty: 4},
	})
	if total != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want zero and %v", total, err, stock.ErrUnknownItem)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 10 {
		t.Fatalf("stock = (%d apples, %d pears), want (8, 10)", apples, pears)
	}
}
