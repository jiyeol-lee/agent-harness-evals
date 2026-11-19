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

func TestOrderPricingCharacterization(t *testing.T) {
	items := []Item{
		{Name: "apple", UnitPrice: 1e16, Qty: 1},
		{Name: "pear", UnitPrice: 0.5, Qty: 2},
		{Name: "apple", UnitPrice: 1, Qty: 1},
	}
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		published = append(published, e)
	})
	svc := New(st, bus)

	const wantTotal = 1e16
	if total, err := svc.Quote(items); err != nil || total != wantTotal {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", total, err, wantTotal)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 10 || pears != 10 || len(published) != 0 {
		t.Fatalf("Quote changed stock or published an event: stock = (%d, %d), events = %v", apples, pears, published)
	}

	if total, err := svc.Place("o1", items); err != nil || total != wantTotal {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", total, err, wantTotal)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 8 {
		t.Fatalf("stock = (%d, %d), want (8, 8)", apples, pears)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 2, "pear": 2}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Fatalf("events = %v, want %v", published, wantEvents)
	}
}

func TestOrderValidationCharacterization(t *testing.T) {
	for _, tt := range []struct {
		name  string
		items []Item
	}{
		{name: "empty"},
		{name: "zero quantity", items: []Item{{Name: "apple", Qty: 1}, {Name: "apple", Qty: 0}}},
		{name: "negative price", items: []Item{{Name: "apple", Qty: 1}, {Name: "apple", UnitPrice: -1, Qty: 1}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			bus := &events.Bus{}
			bus.Subscribe(func(e events.OrderPlaced) {
				t.Errorf("published event for invalid order: %v", e)
			})
			svc := New(st, bus)

			if total, err := svc.Quote(tt.items); err == nil || total != 0 {
				t.Errorf("Quote = (%v, %v), want zero total and an error", total, err)
			}
			if total, err := svc.Place("o1", tt.items); err == nil || total != 0 {
				t.Errorf("Place = (%v, %v), want zero total and an error", total, err)
			}
			if left := st.Count("apple"); left != 10 {
				t.Errorf("apple left = %d, want 10", left)
			}
		})
	}
}

func TestPlaceReservationFailureCharacterization(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("published event for failed reservation: %v", e)
	})
	svc := New(st, bus)
	items := []Item{{Name: "apple", Qty: 1}, {Name: "unknown", Qty: 1}, {Name: "pear", Qty: 1}}

	if total, err := svc.Place("o1", items); !errors.Is(err, stock.ErrUnknownItem) || total != 0 {
		t.Fatalf("Place = (%v, %v), want (0, wrapped ErrUnknownItem)", total, err)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 9 || pears != 10 {
		t.Fatalf("stock = (%d, %d), want (9, 10)", apples, pears)
	}
}
