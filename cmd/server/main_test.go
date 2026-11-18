package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockTotal(t *testing.T) {
	h := newHandler()
	read := func(path, key string) int {
		t.Helper()
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d, want %d", path, w.Code, http.StatusOK)
		}
		if path == "/stock-total" && w.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("GET %s Content-Type = %q, want application/json", path, w.Header().Get("Content-Type"))
		}
		var body map[string]int
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		value, ok := body[key]
		if !ok {
			t.Fatalf("GET %s: missing %q in %v", path, key, body)
		}
		return value
	}

	initial := read("/stock-total", "total")
	if initial != 150 {
		t.Fatalf("initial total = %d, want 150", initial)
	}

	w := httptest.NewRecorder()
	order := strings.NewReader(`{"id":"o1","items":[{"name":"apple","qty":3,"unitPrice":2}]}`)
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/orders", order))
	if w.Code != http.StatusOK {
		t.Fatalf("POST /orders = %d, want %d: %s", w.Code, http.StatusOK, w.Body)
	}

	total := read("/stock-total", "total")
	want := read("/stock/apple", "count") + read("/stock/pear", "count")
	if total != want || total >= initial {
		t.Fatalf("total after order = %d, want %d and less than %d", total, want, initial)
	}
}
