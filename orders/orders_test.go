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

func TestQuotePricesAllItemsWithoutReserving(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("quote published %+v", e)
	})
	svc := New(st, bus)

	items := []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
	}
	if got, err := svc.Quote(items); err != nil || got != 8.75 {
		t.Fatalf("Quote = (%v, %v), want (8.75, nil)", got, err)
	}
	if got := st.Count("apple"); got != 10 {
		t.Errorf("apple left = %d, want 10", got)
	}
}

func TestPlacePublishesAfterReservingAllItems(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		published = append(published, e)
		if got := st.Count("apple"); got != 7 {
			t.Errorf("apple left at publish = %d, want 7", got)
		}
		if got := st.Count("pear"); got != 7 {
			t.Errorf("pear left at publish = %d, want 7", got)
		}
	})
	svc := New(st, bus)

	items := []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
		{Name: "apple", UnitPrice: 3, Qty: 1},
	}
	if got, err := svc.Place("o1", items); err != nil || got != 11.75 {
		t.Fatalf("Place = (%v, %v), want (11.75, nil)", got, err)
	}
	want := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 3}}}
	if !reflect.DeepEqual(published, want) {
		t.Errorf("published = %+v, want %+v", published, want)
	}
}

func TestInvalidOrdersDoNotReserveOrPublish(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("invalid order published %+v", e)
	})
	svc := New(st, bus)
	valid := Item{Name: "apple", UnitPrice: 2.5, Qty: 2}

	for _, tt := range []struct {
		name  string
		items []Item
	}{
		{name: "empty"},
		{name: "zero quantity", items: []Item{valid, {Name: "apple", UnitPrice: 2.5, Qty: 0}}},
		{name: "negative price", items: []Item{valid, {Name: "apple", UnitPrice: -1, Qty: 1}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := svc.Quote(tt.items); err == nil || got != 0 {
				t.Errorf("Quote = (%v, %v), want zero and a validation error", got, err)
			}
			if got, err := svc.Place("o1", tt.items); err == nil || got != 0 {
				t.Errorf("Place = (%v, %v), want zero and a validation error", got, err)
			}
			if got := st.Count("apple"); got != 10 {
				t.Errorf("apple left = %d, want 10", got)
			}
		})
	}
}

func TestPlaceStopsReservingAtFirstStockError(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("failed reservation published %+v", e)
	})
	svc := New(st, bus)

	items := []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 2, Qty: 1},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
	}
	if got, err := svc.Place("o1", items); !errors.Is(err, stock.ErrUnknownItem) || got != 0 {
		t.Fatalf("Place = (%v, %v), want zero and ErrUnknownItem", got, err)
	}
	if got := st.Count("apple"); got != 8 {
		t.Errorf("apple left = %d, want 8", got)
	}
	if got := st.Count("pear"); got != 10 {
		t.Errorf("pear left = %d, want 10", got)
	}
}
