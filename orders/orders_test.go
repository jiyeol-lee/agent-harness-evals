package orders

import (
	"math"
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

func TestOrderQuoteAndPlaceTotals(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prices []float64
		want   float64
	}{
		{name: "fold in input order", prices: []float64{1e16, 1, 1}, want: 1e16},
		{name: "negative zero sums to positive zero", prices: []float64{math.Copysign(0, -1)}, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := stock.NewStock()
			st.Add("apple", 10)
			svc := New(st, &events.Bus{})
			var items []Item
			for _, price := range tc.prices {
				items = append(items, Item{Name: "apple", UnitPrice: price, Qty: 1})
			}

			quoted, err := svc.Quote(items)
			if err != nil || math.Float64bits(quoted) != math.Float64bits(tc.want) {
				t.Fatalf("Quote = (%v, %v), want (%v, nil)", quoted, err, tc.want)
			}
			placed, err := svc.Place("o1", items)
			if err != nil || math.Float64bits(placed) != math.Float64bits(tc.want) {
				t.Fatalf("Place = (%v, %v), want (%v, nil)", placed, err, tc.want)
			}
		})
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
