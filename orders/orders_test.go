package orders

import (
	"errors"
	"reflect"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestOrderTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("a", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("Quote published %+v", e)
	})
	svc := New(st, bus)
	items := []Item{
		{Name: "a", UnitPrice: 0.1, Qty: 1},
		{Name: "b", UnitPrice: 0.2, Qty: 1},
		{Name: "c", UnitPrice: 0.3, Qty: 1},
	}

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
	if left := st.Count("apple"); left != 7 {
		t.Fatalf("apple left = %d, want 7", left)
	}
	if left := st.Count("pear"); left != 6 {
		t.Fatalf("pear left = %d, want 6", left)
	}
	want := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 4}}}
	if !reflect.DeepEqual(published, want) {
		t.Fatalf("published = %+v, want %+v", published, want)
	}
}

func TestOrderValidationBeforeReservation(t *testing.T) {
	valid := Item{Name: "apple", UnitPrice: 2.5, Qty: 2}
	for _, tc := range []struct {
		name  string
		items []Item
	}{
		{name: "empty"},
		{name: "zero quantity", items: []Item{valid, {Name: "apple", UnitPrice: 1}}},
		{name: "negative quantity", items: []Item{valid, {Name: "apple", UnitPrice: 1, Qty: -1}}},
		{name: "negative price", items: []Item{valid, {Name: "apple", UnitPrice: -1, Qty: 1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			bus := &events.Bus{}
			bus.Subscribe(func(e events.OrderPlaced) {
				t.Errorf("invalid order published %+v", e)
			})
			svc := New(st, bus)

			if total, err := svc.Quote(tc.items); total != 0 || err == nil {
				t.Errorf("Quote = (%v, %v), want (0, error)", total, err)
			}
			if total, err := svc.Place("invalid", tc.items); total != 0 || err == nil {
				t.Errorf("Place = (%v, %v), want (0, error)", total, err)
			}
			if left := st.Count("apple"); left != 10 {
				t.Errorf("apple left = %d, want 10", left)
			}
		})
	}
}

func TestPlaceReservationFailure(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("failed order published %+v", e)
	})
	svc := New(st, bus)

	total, err := svc.Place("failed", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "unknown", UnitPrice: 1, Qty: 1},
	})
	if total != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want (0, wrapped ErrUnknownItem)", total, err)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}
