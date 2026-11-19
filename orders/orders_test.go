package orders

import (
	"errors"
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

func TestQuotePreservesItemOrderWithoutReservation(t *testing.T) {
	items := []Item{
		{Name: "apple", UnitPrice: 0.1, Qty: 1},
		{Name: "pear", UnitPrice: 0.2, Qty: 1},
		{Name: "plum", UnitPrice: 0.3, Qty: 1},
	}
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(events.OrderPlaced) { t.Error("quote published an order") })
	svc := New(st, bus)

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.6000000000000001; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left = %d, want 10", left)
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 7 || pears != 7 {
			t.Errorf("stock at publication: apple = %d, pear = %d, want 7 each", apples, pears)
		}
		published = append(published, e)
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
		{Name: "apple", UnitPrice: 2.5, Qty: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := 11.25; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if len(published) != 1 {
		t.Fatalf("published %d events, want 1", len(published))
	}
	e := published[0]
	if e.OrderID != "o1" || len(e.Items) != 2 || e.Items["apple"] != 3 || e.Items["pear"] != 3 {
		t.Errorf("published unexpected order: %+v", e)
	}
}

func TestOrderValidation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		items []Item
		want  string
	}{
		{name: "empty", want: "order has no items"},
		{
			name: "quantity before price",
			items: []Item{
				{Name: "apple", UnitPrice: 2.5, Qty: 2},
				{Name: "pear", UnitPrice: -1, Qty: 0},
			},
			want: "item pear: quantity must be positive",
		},
		{
			name:  "negative price",
			items: []Item{{Name: "apple", UnitPrice: -1, Qty: 2}},
			want:  "item apple: price must not be negative",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			bus := &events.Bus{}
			bus.Subscribe(func(events.OrderPlaced) { t.Error("invalid order was published") })
			svc := New(st, bus)

			if total, err := svc.Quote(tc.items); total != 0 || err == nil || err.Error() != tc.want {
				t.Errorf("Quote = (%v, %v), want (0, %q)", total, err, tc.want)
			}
			if total, err := svc.Place("o1", tc.items); total != 0 || err == nil || err.Error() != tc.want {
				t.Errorf("Place = (%v, %v), want (0, %q)", total, err, tc.want)
			}
			if left := st.Count("apple"); left != 10 {
				t.Errorf("apple left = %d, want 10", left)
			}
		})
	}
}

func TestPlaceStopsAtFailedReservation(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(events.OrderPlaced) { t.Error("failed order was published") })
	svc := New(st, bus)

	total, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "missing", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
	})
	if total != 0 || !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("Place = (%v, %v), want (0, wrapped ErrUnknownItem)", total, err)
	}
	if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 10 {
		t.Errorf("stock after failure: apple = %d, pear = %d, want 8 and 10", apples, pears)
	}
}
