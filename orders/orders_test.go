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
	st.Add("bulk", 2)
	st.Add("extra", 2)
	svc := New(st, &events.Bus{})

	// Adding the two extras separately leaves the total unchanged; adding
	// them together first would increase it by two.
	items := []Item{
		{Name: "bulk", UnitPrice: 4503599627370496, Qty: 2},
		{Name: "extra", UnitPrice: 1, Qty: 1},
		{Name: "extra", UnitPrice: 1, Qty: 1},
	}
	const want = 9007199254740992.0

	quoted, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if quoted != want {
		t.Errorf("quote = %v, want %v", quoted, want)
	}

	placed, err := svc.Place("o1", items)
	if err != nil {
		t.Fatal(err)
	}
	if placed != want {
		t.Errorf("placed total = %v, want %v", placed, want)
	}
}
