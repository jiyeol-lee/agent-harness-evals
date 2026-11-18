package orders

import (
	"errors"
	"reflect"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestOrderTotal(t *testing.T) {
	items := []Item{
		{Name: "a", UnitPrice: 0.125, Qty: 2},
		{Name: "b", UnitPrice: 0.25, Qty: 1},
		{Name: "c", UnitPrice: 0.5, Qty: 3},
	}
	svc := New(stock.NewStock(), &events.Bus{})

	got, err := svc.Quote(items)
	if err != nil {
		t.Fatal(err)
	}
	if want := 2.0; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 5)
	bus := &events.Bus{}
	var published []events.OrderPlaced
	bus.Subscribe(func(e events.OrderPlaced) {
		published = append(published, e)
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
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
	if left := st.Count("pear"); left != 2 {
		t.Fatalf("pear left = %d, want 2", left)
	}
	wantEvents := []events.OrderPlaced{{OrderID: "o1", Items: map[string]int{"apple": 2, "pear": 3}}}
	if !reflect.DeepEqual(published, wantEvents) {
		t.Fatalf("published = %v, want %v", published, wantEvents)
	}
}

func TestPlaceValidatesBeforeReserving(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	svc := New(st, &events.Bus{})

	_, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: -1, Qty: 1},
	})
	if err == nil || errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("error = %v, want validation error", err)
	}
	if left := st.Count("apple"); left != 10 {
		t.Fatalf("apple left = %d, want 10", left)
	}
}

func TestPlaceDoesNotPublishAfterReservationFailure(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		t.Errorf("published after reservation failure: %v", e)
	})
	svc := New(st, bus)

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.25, Qty: 1},
	})
	if !errors.Is(err, stock.ErrUnknownItem) {
		t.Fatalf("error = %v, want %v", err, stock.ErrUnknownItem)
	}
	if got != 0 {
		t.Fatalf("total = %v, want 0", got)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
}
