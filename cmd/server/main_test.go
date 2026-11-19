package main

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockRoutes(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 5)
	st.Add("pear", 3)
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	checkJSON := func(t *testing.T, path, key string, value int) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s: status = %d, want %d", path, rec.Code, http.StatusOK)
		}
		var got map[string]int
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("GET %s: invalid JSON: %v", path, err)
		}
		want := map[string]int{key: value}
		if !maps.Equal(got, want) {
			t.Errorf("GET %s: body = %v, want %v", path, got, want)
		}
		if path == "/stock-total" && rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("GET %s: Content-Type = %q, want application/json", path, rec.Header().Get("Content-Type"))
		}
	}

	t.Run("initial total", func(t *testing.T) {
		checkJSON(t, "/stock-total", "total", 8)
	})
	t.Run("updated inventory", func(t *testing.T) {
		if err := st.Remove("apple", 2); err != nil {
			t.Fatal(err)
		}
		checkJSON(t, "/stock-total", "total", 6)
	})
	t.Run("POST not allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/stock-total", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST /stock-total: status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
		}
	})
	t.Run("item count", func(t *testing.T) {
		checkJSON(t, "/stock/pear", "count", 3)
	})
}
