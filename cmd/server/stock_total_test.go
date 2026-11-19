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

	checkTotal := func(t *testing.T, want int) {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stock-total", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
		if got := w.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var body map[string]*int
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode stock total: %v", err)
		}
		if total := body["total"]; len(body) != 1 || total == nil || *total != want {
			t.Errorf("body = %s, want {\"total\":%d}", w.Body.Bytes(), want)
		}
	}

	t.Run("empty", func(t *testing.T) { checkTotal(t, 0) })
	st.Add("apple", 3)
	st.Add("pear", 5)
	t.Run("populated", func(t *testing.T) { checkTotal(t, 8) })
	if err := st.Remove("apple", 2); err != nil {
		t.Fatalf("Remove(apple, 2): %v", err)
	}
	t.Run("after removal", func(t *testing.T) { checkTotal(t, 6) })

	t.Run("wrong method", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/stock-total", nil))
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
		}
	})
}
