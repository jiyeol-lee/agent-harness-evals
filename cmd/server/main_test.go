package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/shop/stock"
)

func TestStockRoutes(t *testing.T) {
	stocked := stock.NewStock()
	stocked.Add("pear", 3)
	stocked.Add("banana", 2)
	stocked.Add("apple", 99)
	stocked.Add("apple", 1)
	if err := stocked.Remove("banana", 2); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name string
		st   *stock.Stock
		path string
		want string
	}{
		{"list", stocked, "/stock", `[{"name":"apple","count":100},{"name":"banana","count":0},{"name":"pear","count":3}]`},
		{"empty list", stock.NewStock(), "/stock", `[]`},
		{"per-name count", stocked, "/stock/apple", `{"count":100}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			registerStockRoutes(mux, tt.st)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if w.Code != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want 200", tt.path, w.Code)
			}
			if got := strings.TrimSpace(w.Body.String()); got != tt.want {
				t.Errorf("GET %s: body = %s, want %s", tt.path, got, tt.want)
			}
		})
	}
}
