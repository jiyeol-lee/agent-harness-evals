package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockEndpoints(t *testing.T) {
	st := stock.NewStock()
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	check := func(path, want string) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Fatalf("GET %s: status %d, body %q; want 200, %q", path, rec.Code, rec.Body.String(), want)
		}
		if path == "/stock" && rec.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("GET /stock Content-Type = %q", rec.Header().Get("Content-Type"))
		}
	}

	check("/stock", "[]\n")
	st.Add("pear", -2)
	st.Add("apple", 0)
	check("/stock", "[{\"name\":\"apple\",\"count\":0},{\"name\":\"pear\",\"count\":-2}]\n")
	check("/stock/pear", "{\"count\":-2}\n")
}
