package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestStockRoutes(t *testing.T) {
	st := stock.NewStock()
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	t.Run("empty list", func(t *testing.T) {
		checkStockResponse(t, mux, "/stock", `[]`, true)
	})

	st.Add("pear", 2)
	st.Add("apple", 5)
	st.Add("apple", 2)
	if err := st.Remove("pear", 2); err != nil {
		t.Fatal(err)
	}

	t.Run("sorted current counts", func(t *testing.T) {
		checkStockResponse(t, mux, "/stock", `[{"name":"apple","count":7},{"name":"pear","count":0}]`, true)
	})
	t.Run("named lookup", func(t *testing.T) {
		checkStockResponse(t, mux, "/stock/apple", `{"count":7}`, false)
	})
	t.Run("unknown name", func(t *testing.T) {
		checkStockResponse(t, mux, "/stock/orange", `{"count":0}`, false)
	})
}

func TestStockRoutesConcurrentUpdates(t *testing.T) {
	st := stock.NewStock()
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 1; i <= 100; i++ {
			st.Add("apple", i)
			if err := st.Remove("apple", i); err != nil {
				t.Error(err)
			}
		}
	}()
	defer func() { <-done }()

	list := httptest.NewRequest(http.MethodGet, "/stock", nil)
	named := httptest.NewRequest(http.MethodGet, "/stock/apple", nil)
	for range 100 {
		mux.ServeHTTP(httptest.NewRecorder(), list)
		mux.ServeHTTP(httptest.NewRecorder(), named)
	}
}

func checkStockResponse(t *testing.T, mux *http.ServeMux, path, wantJSON string, requireJSONHeader bool) {
	t.Helper()
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want %d", path, w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Content-Type"); requireJSONHeader && got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got, want any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if err := json.Unmarshal([]byte(wantJSON), &want); err != nil {
		t.Fatalf("decode expected JSON: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("body = %s, want %s", w.Body.Bytes(), wantJSON)
	}
}
