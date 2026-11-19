package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"example.com/shop/stock"
)

func TestStockList(t *testing.T) {
	st := stock.NewStock()
	st.Add("pear", 5)
	st.Add("orange", 1)
	st.Add("apple", 2)
	st.Add("pear", 3)
	st.Add("banana", -1)
	if err := st.Remove("orange", 1); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		st   *stock.Stock
		want []map[string]any
	}{
		{
			name: "current counts sorted by name",
			st:   st,
			want: []map[string]any{
				{"name": "apple", "count": float64(2)},
				{"name": "banana", "count": float64(-1)},
				{"name": "orange", "count": float64(0)},
				{"name": "pear", "count": float64(8)},
			},
		},
		{
			name: "empty stock",
			st:   stock.NewStock(),
			want: []map[string]any{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			registerStockRoutes(mux, tc.st)
			res := httptest.NewRecorder()
			mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))

			if res.Code != http.StatusOK {
				t.Fatalf("GET /stock status = %d, want %d; body = %q", res.Code, http.StatusOK, res.Body.String())
			}
			if got := res.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("GET /stock Content-Type = %q, want %q", got, "application/json")
			}
			var got []map[string]any
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("GET /stock body = %s, want %v", res.Body.String(), tc.want)
			}
		})
	}
}

func TestStockListDuringUpdates(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 100)
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	var updates sync.WaitGroup
	updates.Go(func() {
		for range 100 {
			st.Add("apple", 1)
			_ = st.Remove("apple", 2)
		}
	})
	defer updates.Wait()

	for range 100 {
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))
		if res.Code != http.StatusOK {
			t.Fatalf("GET /stock status = %d, want %d", res.Code, http.StatusOK)
		}
	}
}
