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
	mux := http.NewServeMux()
	registerStockTotal(mux, st)

	assertTotal := func(t *testing.T, want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock-total", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /stock-total status = %d, want %d", rec.Code, http.StatusOK)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var body map[string]*int
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if total := body["total"]; len(body) != 1 || total == nil || *total != want {
			t.Errorf("GET /stock-total body = %q, want JSON object with only total = %d", rec.Body.String(), want)
		}
	}

	t.Run("fresh stock", func(t *testing.T) {
		assertTotal(t, 0)
	})

	st.Add("apple", 3)
	st.Add("pear", 4)
	t.Run("populated stock", func(t *testing.T) {
		assertTotal(t, 7)
	})

	st.Add("apple", 2)
	if err := st.Remove("pear", 1); err != nil {
		t.Fatalf("Remove(pear, 1): %v", err)
	}
	t.Run("after mutations", func(t *testing.T) {
		assertTotal(t, 8)
	})
}

func TestStockTotalRejectsPOST(t *testing.T) {
	mux := http.NewServeMux()
	registerStockTotal(mux, stock.NewStock())

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/stock-total", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /stock-total status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
