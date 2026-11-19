package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestStockList(t *testing.T) {
	populated := stock.NewStock()
	populated.Add("pear", 2)
	populated.Add("banana", 1)
	populated.Add("apple", 3)
	populated.Add("pear", 4)
	if err := populated.Remove("banana", 1); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		st   *stock.Stock
		want []map[string]any
	}{
		{
			name: "sorted with current counts including zero",
			st:   populated,
			want: []map[string]any{
				{"name": "apple", "count": float64(3)},
				{"name": "banana", "count": float64(0)},
				{"name": "pear", "count": float64(6)},
			},
		},
		{
			name: "empty array",
			st:   stock.NewStock(),
			want: []map[string]any{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			registerStockRoutes(mux, tt.st)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/stock", nil))

			if got := response.Code; got != http.StatusOK {
				t.Fatalf("GET /stock status = %d, want %d", got, http.StatusOK)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("GET /stock Content-Type = %q, want application/json", got)
			}
			var got []map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("GET /stock invalid JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GET /stock body = %s, want %v", response.Body, tt.want)
			}
		})
	}
}

func TestNamedStock(t *testing.T) {
	st := stock.NewStock()
	st.Add("pear", 2)
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	for _, tt := range []struct {
		path string
		want int
	}{
		{path: "/stock/pear", want: 2},
		{path: "/stock/unknown", want: 0},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

		if got := response.Code; got != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", tt.path, got, http.StatusOK)
		}
		var got map[string]int
		if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatalf("GET %s invalid JSON: %v", tt.path, err)
		}
		want := map[string]int{"count": tt.want}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GET %s body = %s, want %v", tt.path, response.Body, want)
		}
	}
}
