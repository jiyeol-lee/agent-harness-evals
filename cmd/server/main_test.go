package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockRoutes(t *testing.T) {
	s := stock.NewStock()
	mux := http.NewServeMux()
	registerStockRoutes(mux, s)

	check := func(path, body string) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != body {
			t.Fatalf("GET %s: status %d, body %q; want 200, %q", path, rec.Code, rec.Body.String(), body)
		}
		if path == "/stock" && rec.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("GET /stock Content-Type = %q, want application/json", rec.Header().Get("Content-Type"))
		}
	}

	check("/stock", "[]\n")
	s.Add("pear", 0)
	s.Add("apple", 3)
	s.Add("banana", -1)
	check("/stock", "[{\"name\":\"apple\",\"count\":3},{\"name\":\"banana\",\"count\":-1},{\"name\":\"pear\",\"count\":0}]\n")
	check("/stock/apple", "{\"count\":3}\n")
	check("/stock/missing", "{\"count\":0}\n")
}
