package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/stock"
)

func TestGetStock(t *testing.T) {
	populated := stock.NewStock()
	populated.Add("pear", 4)
	populated.Add("apple", 2)
	populated.Add("apple", 3)
	if err := populated.Remove("pear", 4); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		st   *stock.Stock
		want []map[string]any
	}{
		{
			name: "sorted current counts including zero",
			st:   populated,
			want: []map[string]any{
				{"name": "apple", "count": float64(5)},
				{"name": "pear", "count": float64(0)},
			},
		},
		{
			name: "empty array",
			st:   stock.NewStock(),
			want: []map[string]any{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mux := newMux(tt.st, orders.New(tt.st, &events.Bus{}))
			res := httptest.NewRecorder()
			mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))

			if res.Code != http.StatusOK {
				t.Fatalf("GET /stock status = %d, want %d", res.Code, http.StatusOK)
			}
			if got := res.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var got []map[string]any
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GET /stock body = %s, want %#v", res.Body.Bytes(), tt.want)
			}
		})
	}
}
