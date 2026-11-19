package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	st.Add("pear", 5)
	mux := http.NewServeMux()
	registerStockTotal(mux, st)

	assertTotal := func(want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock-total", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /stock-total status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		var body map[string]int
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode stock total: %v", err)
		}
		if len(body) != 1 || body["total"] != want {
			t.Fatalf("response = %s, want {\"total\": %d}", rec.Body, want)
		}
	}

	assertTotal(8)
	if err := st.Remove("apple", 2); err != nil {
		t.Fatal(err)
	}
	assertTotal(6)
}

func TestStockTotalRejectsPost(t *testing.T) {
	mux := http.NewServeMux()
	registerStockTotal(mux, stock.NewStock())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/stock-total", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /stock-total status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
