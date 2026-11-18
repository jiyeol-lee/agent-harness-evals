package orders

import (
	"reflect"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestOrderTotal(t *testing.T) {
	items := []Item{
		{Name: "a", UnitPrice: 0.1, Qty: 1},
		{Name: "b", UnitPrice: 0.2, Qty: 1},
		{Name: "c", UnitPrice: 0.3, Qty: 1},
	}
	svc := New(stock.NewStock(), &events.Bus{})

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 0.6000000000000001; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 10)
	bus := &events.Bus{}
	var placed []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		placed = append(placed, e)
		if apples, pears := st.Count("apple"), st.Count("pear"); apples != 8 || pears != 7 {
			t.Errorf("stock at publication = (%d apples, %d pears), want (8, 7)", apples, pears)
		}
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := 8.75; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	wantPlaced := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 2, "pear": 3}}}
	if !reflect.DeepEqual(placed, wantPlaced) {
		t.Fatalf("published events = %v, want %v", placed, wantPlaced)
	}
}
