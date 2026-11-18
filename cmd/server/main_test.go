package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockTotal(t *testing.T) {
	handler := newHandler()
	if got := readStock(t, handler, "/stock-total", "total"); got != 150 {
		t.Fatalf("total = %d, want 150", got)
	}
}

func TestStockTotalAfterOrder(t *testing.T) {
	handler := newHandler()
	placeAppleOrder(t, handler)

	apple := readStock(t, handler, "/stock/apple", "count")
	// The existing 10ms dedupe may suppress the subscriber's second removal.
	if apple != 98 && apple != 96 {
		t.Fatalf("apple count = %d, want 98 or 96", apple)
	}
	if got, want := readStock(t, handler, "/stock-total", "total"), apple+50; got != want {
		t.Fatalf("total = %d, want %d", got, want)
	}
}

func TestStockTotalConcurrentRequests(t *testing.T) {
	handler := newHandler()
	t.Run("order", func(t *testing.T) {
		t.Parallel()
		placeAppleOrder(t, handler)
	})
	t.Run("total", func(t *testing.T) {
		t.Parallel()
		readStock(t, handler, "/stock-total", "total")
	})
	t.Run("count", func(t *testing.T) {
		t.Parallel()
		readStock(t, handler, "/stock/apple", "count")
	})
}

func placeAppleOrder(t *testing.T, handler http.Handler) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/orders",
		strings.NewReader(`{"id":"o1","items":[{"name":"apple","unitPrice":2.5,"qty":2}]}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /orders: status = %d, want 200; body = %s", rec.Code, rec.Body)
	}
}

func readStock(t *testing.T, handler http.Handler, path, key string) int {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want 200; body = %s", path, rec.Code, rec.Body)
	}
	var body map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("GET %s: decode JSON: %v", path, err)
	}
	value, ok := body[key]
	if !ok || len(body) != 1 {
		t.Fatalf("GET %s: body = %s, want an object with only %q", path, rec.Body, key)
	}
	return value
}
