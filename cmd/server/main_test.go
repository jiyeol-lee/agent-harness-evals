package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"example.com/shop/stock"
)

const appleOrder = `{"id":"order","items":[{"name":"apple","unitPrice":1,"qty":1}]}`

func TestStockTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 10)
	st.Add("pear", 5)
	handler := newHandler(st)

	total := func(t *testing.T) int {
		t.Helper()
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock-total", nil))
		if res.Code != http.StatusOK {
			t.Fatalf("GET /stock-total status = %d, want 200", res.Code)
		}
		if got := res.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}
		var body map[string]int
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode stock total: %v", err)
		}
		got, ok := body["total"]
		if len(body) != 1 || !ok {
			t.Fatalf("unexpected stock total response: %s", res.Body)
		}
		return got
	}

	t.Run("initial", func(t *testing.T) {
		if got := total(t); got != 15 {
			t.Errorf("total = %d, want 15", got)
		}
	})

	t.Run("after order", func(t *testing.T) {
		order := httptest.NewRecorder()
		handler.ServeHTTP(order, httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(appleOrder)))
		if order.Code != http.StatusOK {
			t.Fatalf("POST /orders status = %d, want 200: %s", order.Code, order.Body)
		}

		item := httptest.NewRecorder()
		handler.ServeHTTP(item, httptest.NewRequest(http.MethodGet, "/stock/apple", nil))
		if item.Code != http.StatusOK {
			t.Fatalf("GET /stock/apple status = %d, want 200", item.Code)
		}
		var body struct{ Count *int }
		if err := json.Unmarshal(item.Body.Bytes(), &body); err != nil || body.Count == nil {
			t.Fatalf("invalid stock count response: %s", item.Body)
		}
		if got, want := total(t), *body.Count+5; got != want || got >= 15 {
			t.Errorf("total = %d, want %d and less than the initial total of 15", got, want)
		}
	})

	t.Run("rejects POST", func(t *testing.T) {
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/stock-total", nil))
		if res.Code != http.StatusMethodNotAllowed {
			t.Errorf("POST /stock-total status = %d, want 405", res.Code)
		}
	})
}

func TestConcurrentStockRequests(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 100)
	handler := newHandler(st)

	start := make(chan struct{})
	var requests sync.WaitGroup
	for range 10 {
		for _, route := range []struct{ method, path, body string }{
			{http.MethodPost, "/orders", appleOrder},
			{http.MethodGet, "/stock/apple", ""},
			{http.MethodGet, "/stock-total", ""},
		} {
			requests.Go(func() {
				<-start
				res := httptest.NewRecorder()
				handler.ServeHTTP(res, httptest.NewRequest(route.method, route.path, strings.NewReader(route.body)))
				if res.Code != http.StatusOK {
					t.Errorf("%s %s status = %d, want 200", route.method, route.path, res.Code)
				}
			})
		}
	}
	close(start)
	requests.Wait()
}
