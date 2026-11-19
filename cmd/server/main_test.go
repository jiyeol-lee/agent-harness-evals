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
	for _, tt := range []struct {
		name     string
		addItems bool
		want     []map[string]any
	}{
		{
			name:     "sorted including zero",
			addItems: true,
			want: []map[string]any{
				{"name": "apple", "count": float64(0)},
				{"name": "pear", "count": float64(5)},
			},
		},
		{
			name: "empty array",
			want: []map[string]any{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			if tt.addItems {
				st.Add("pear", 5)
				st.Add("apple", 2)
				if err := st.Remove("apple", 2); err != nil {
					t.Fatal(err)
				}
			}
			mux := newMux(st, orders.New(st, &events.Bus{}))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("GET /stock status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var got []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("body = %s, want %#v", rec.Body.Bytes(), tt.want)
			}
		})
	}
}
