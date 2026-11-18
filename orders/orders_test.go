package orders

import (
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

func TestMultiLineOrderTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	st.Add("pear", 1)
	svc := New(st, &events.Bus{})
	items := []Item{
		{Name: "apple", UnitPrice: 0.5, Qty: 2},
		{Name: "apple", UnitPrice: 1, Qty: 1},
		{Name: "pear", UnitPrice: 1e16, Qty: 1},
	}
	// Adding the two small lines first preserves their contribution.
	const want = 10000000000000002

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Quote total = %v, want %v", got, float64(want))
	}

	got, err = svc.Place("o1", items)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Place total = %v, want %v", got, float64(want))
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
