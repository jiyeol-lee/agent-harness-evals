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
	st.Add("plum", 10)
	svc := New(st, &events.Bus{})
	items := []Item{
		{Name: "apple", UnitPrice: 0.1, Qty: 1},
		{Name: "pear", UnitPrice: 0.2, Qty: 1},
		{Name: "plum", UnitPrice: 0.3, Qty: 1},
	}
	const want = 0.6000000000000001

	if got, err := svc.Quote(items); err != nil || got != want {
		t.Fatalf("Quote = (%v, %v), want (%v, nil)", got, err, want)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left after quote = %d, want 10", left)
	}
	if got, err := svc.Place("o1", items); err != nil || got != want {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", got, err, want)
	}
}

func TestPlaceSumsRoundedLineQuotes(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 4)
	svc := New(st, &events.Bus{})

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 0.01, Qty: 1},
		{Name: "apple", UnitPrice: 0.10, Qty: 3},
	})
	const want = 0.31000000000000005
	if err != nil || got != want {
		t.Fatalf("Place = (%v, %v), want (%v, nil)", got, err, want)
	}
}

func TestPlaceValidatesBeforeReserving(t *testing.T) {
	tests := []struct {
		name string
		item Item
	}{
		{name: "zero quantity", item: Item{Name: "pear", UnitPrice: 1, Qty: 0}},
		{name: "negative price", item: Item{Name: "pear", UnitPrice: -1, Qty: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			st.Add("pear", 10)
			svc := New(st, &events.Bus{})
			items := []Item{{Name: "apple", UnitPrice: 1, Qty: 2}, tt.item}

			if _, err := svc.Quote(items); err == nil {
				t.Fatal("Quote succeeded with an invalid item")
			}
			if _, err := svc.Place("o1", items); err == nil {
				t.Fatal("Place succeeded with an invalid item")
			}
			if left := st.Count("apple"); left != 10 {
				t.Fatalf("apple left = %d, want 10", left)
			}
		})
	}
}

func TestPlaceReservationFailureKeepsEarlierReservation(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(events.OrderPlaced) {
		t.Error("published OrderPlaced after failed reservation")
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 1, Qty: 2},
		{Name: "missing", UnitPrice: 1, Qty: 1},
	})
	if !errors.Is(err, stock.ErrUnknownItem) || got != 0 {
		t.Fatalf("Place = (%v, %v), want (0, ErrUnknownItem)", got, err)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}
