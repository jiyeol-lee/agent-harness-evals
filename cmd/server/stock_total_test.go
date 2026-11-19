package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotalEmpty(t *testing.T) {
	mux := http.NewServeMux()
	registerStockTotal(mux, stock.NewStock())

	assertStockTotal(t, mux, 0)
}

func TestStockTotalTracksCurrentQuantities(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	st.Add("pear", 2)
	mux := http.NewServeMux()
	registerStockTotal(mux, st)

	assertStockTotal(t, mux, 5)

	st.Add("apple", 4)
	if err := st.Remove("pear", 1); err != nil {
		t.Fatal(err)
	}
	assertStockTotal(t, mux, 8)
}

func TestStockTotalRejectsPost(t *testing.T) {
	mux := http.NewServeMux()
	registerStockTotal(mux, stock.NewStock())

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/stock-total", nil))

	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /stock-total status = %d, want 405", res.Code)
	}
}

func assertStockTotal(t *testing.T, mux *http.ServeMux, want int) {
	t.Helper()

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock-total", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("GET /stock-total status = %d, want 200", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("GET /stock-total Content-Type = %q, want application/json", got)
	}
	var body map[string]*int
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET /stock-total body = %q, want JSON object with an integer total: %v", res.Body, err)
	}
	if got := body["total"]; got == nil || *got != want || len(body) != 1 {
		t.Fatalf("GET /stock-total body = %s, want {\"total\":%d}", res.Body, want)
	}
}
