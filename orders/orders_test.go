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

func TestQuoteAndPlaceUseInputOrder(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		published = append(published, e)
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 7 || pears != 9 {
			t.Errorf("stock at publish = (%d, %d), want (7, 9)", apples, pears)
		}
	})
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 0.05, Qty: 2},
		{Name: "pear", UnitPrice: 0.2, Qty: 1},
		{Name: "apple", UnitPrice: 0.3, Qty: 1},
	}
	const want = 0.6000000000000001
	if total, err := svc.Quote(items); err != nil || total != want {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", total, err, want)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 10 || pears != 10 || len(published) != 0 {
		t.Fatalf("after Quote: stock = (%d, %d), events = %v; want unchanged stock and no events", apples, pears, published)
	}
	if total, err := svc.Place("o1", items); err != nil || total != want {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", total, err, want)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 3, "pear": 1}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Fatalf("events = %v, want %v", published, wantEvents)
	}
}

func TestQuoteAndPlaceRejectInvalidOrdersBeforeReserving(t *testing.T) {
	valid := Item{Name: "apple", UnitPrice: 2.5, Qty: 2}
	tests := []struct {
		name  string
		items []Item
		want  string
	}{
		{"empty", nil, "order has no items"},
		{"quantity before price", []Item{valid, {Name: "pear", UnitPrice: -1}}, "item pear: quantity must be positive"},
		{"negative price", []Item{valid, {Name: "pear", UnitPrice: -1, Qty: 1}}, "item pear: price must not be negative"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			bus := &events.Bus{}
			bus.Subscribe(func(e events.OrderPlaced) { t.Errorf("unexpected event: %v", e) })
			svc := New(st, bus)

			for name, call := range map[string]func() (float64, error){
				"Quote": func() (float64, error) { return svc.Quote(tt.items) },
				"Place": func() (float64, error) { return svc.Place("o1", tt.items) },
			} {
				if total, err := call(); total != 0 || err == nil || err.Error() != tt.want {
					t.Errorf("%s = (%v, %v), want (0, %q)", name, total, err, tt.want)
				}
			}
			if apples := st.Count("apple"); apples != 10 {
				t.Errorf("apple stock = %d, want 10", apples)
			}
		})
	}
}

func TestPlaceKeepsEarlierReservationsOnStockError(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) { t.Errorf("unexpected event: %v", e) })
	svc := New(st, bus)
	items := []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "missing", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: 1, Qty: 1},
	}

	if total, err := svc.Place("o1", items); total != 0 || !errors.Is(err, stock.ErrUnknownItem) || err.Error() != "reserve missing: unknown item" {
		t.Fatalf("Place = (%v, %v), want (0, reserve missing: unknown item)", total, err)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 10 {
		t.Errorf("stock after failure = (%d, %d), want (8, 10)", apples, pears)
	}
}
