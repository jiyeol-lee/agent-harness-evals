//go:build gc

package orders

import (
	"math"
	"runtime/debug"
	"testing"

	"example.com/shop/events"
	"example.com/shop/stock"
)

func TestPlacePreservesNaNPayload(t *testing.T) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "-gcflags" && setting.Value != "" {
				t.Skip("baseline bits were measured with default compiler flags")
			}
		}
	}
	st := stock.NewStock()
	st.Add("apple", 3)
	svc := New(st, &events.Bus{})

	got, err := svc.Place("o1", []Item{
		{Name: "apple", UnitPrice: math.Float64frombits(0xfff8000000000001), Qty: 1},
		{Name: "apple", UnitPrice: math.Float64frombits(0x7ff0000000000002), Qty: 2},
	})
	// NaN payload selection is a property of the optimized amd64 build, not of Go.
	const want = 0x7ff8000000000002
	if bits := math.Float64bits(got); err != nil || bits != want {
		t.Fatalf("Place = (%016x, %v), want (%016x, nil)", bits, err, want)
	}
}
