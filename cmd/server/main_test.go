package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockTotal(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock-total", nil))

	if got := rec.Code; got != http.StatusOK {
		t.Fatalf("GET /stock-total status = %d, want %d", got, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("GET /stock-total Content-Type = %q, want application/json", got)
	}
	if got := jsonField(t, rec, "total"); got != 150 {
		t.Fatalf("total = %d, want 150", got)
	}
}

func TestStockTotalAfterOrder(t *testing.T) {
	handler := newHandler()
	order := httptest.NewRecorder()
	handler.ServeHTTP(order, httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(
		`{"id":"o1","items":[{"name":"apple","unitPrice":2,"qty":3}]}`,
	)))
	if got := order.Code; got != http.StatusOK {
		t.Fatalf("POST /orders status = %d, want %d: %s", got, http.StatusOK, order.Body)
	}

	apple := httptest.NewRecorder()
	handler.ServeHTTP(apple, httptest.NewRequest(http.MethodGet, "/stock/apple", nil))
	if got := apple.Code; got != http.StatusOK {
		t.Fatalf("GET /stock/apple status = %d, want %d", got, http.StatusOK)
	}
	count := jsonField(t, apple, "count")
	if count >= 100 {
		t.Fatalf("apple count = %d, want less than the initial 100", count)
	}

	total := httptest.NewRecorder()
	handler.ServeHTTP(total, httptest.NewRequest(http.MethodGet, "/stock-total", nil))
	if got := total.Code; got != http.StatusOK {
		t.Fatalf("GET /stock-total status = %d, want %d", got, http.StatusOK)
	}
	if got, want := jsonField(t, total, "total"), count+50; got != want {
		t.Fatalf("total = %d, want %d (apple count plus 50 pears)", got, want)
	}
}

func TestStockTotalConcurrentRequests(t *testing.T) {
	handler := newHandler()
	for _, route := range []struct {
		method, path, body string
	}{
		{http.MethodPost, "/orders", `{"id":"o1","items":[{"name":"apple","unitPrice":2,"qty":1}]}`},
		{http.MethodGet, "/stock/apple", ""},
		{http.MethodGet, "/stock-total", ""},
	} {
		t.Run(route.path, func(t *testing.T) {
			t.Parallel()
			for range 20 {
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, httptest.NewRequest(route.method, route.path, strings.NewReader(route.body)))
				if got := rec.Code; got != http.StatusOK {
					t.Fatalf("%s %s status = %d, want %d", route.method, route.path, got, http.StatusOK)
				}
			}
		})
	}
}

func jsonField(t *testing.T, rec *httptest.ResponseRecorder, key string) int {
	t.Helper()
	var body map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	value, ok := body[key]
	if !ok {
		t.Fatalf("response %q has no %q field", rec.Body, key)
	}
	return value
}
