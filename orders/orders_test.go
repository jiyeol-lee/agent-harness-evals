package orders

import (
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestOrderTotal(t *testing.T) {
	svc := New(stock.NewStock(), &events.Bus{})
	a := Item{Name: "a", UnitPrice: 0.1, Qty: 1}
	b := Item{Name: "b", UnitPrice: 0.2, Qty: 1}
	c := Item{Name: "c", UnitPrice: 0.3, Qty: 1}
	for _, tt := range []struct {
		name  string
		items []Item
		want  float64
	}{
		{"ascending", []Item{a, b, c}, 0.6000000000000001},
		{"descending", []Item{c, b, a}, 0.6},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Quote(tt.items)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("total = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlaceReservesStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 5)
	svc := New(st, &events.Bus{})

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: 2.5, Qty: 2},
		{Name: "pear", UnitPrice: 1.5, Qty: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := 9.5; got != want {
		t.Fatalf("total = %v, want %v", got, want)
	}
	if left := st.Count("apple"); left != 8 {
		t.Fatalf("apple left = %d, want 8", left)
	}
	if left := st.Count("pear"); left != 2 {
		t.Fatalf("pear left = %d, want 2", left)
	}
}
