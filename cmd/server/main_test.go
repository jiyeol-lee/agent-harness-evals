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
		checkStockList(t, mux, []map[string]any{})
	})

	st.Add("pear", 2)
	st.Add("banana", 0)
	st.Add("apple", 4)
	st.Add("apple", 2)
	if err := st.Remove("apple", 1); err != nil {
		t.Fatal(err)
	}

	t.Run("sorted list with current counts", func(t *testing.T) {
		checkStockList(t, mux, []map[string]any{
			{"name": "apple", "count": float64(5)},
			{"name": "banana", "count": float64(0)},
			{"name": "pear", "count": float64(2)},
		})
	})

	t.Run("single item", func(t *testing.T) {
		resp := httptest.NewRecorder()
		mux.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/stock/apple", nil))
		if resp.Code != http.StatusOK {
			t.Fatalf("GET /stock/apple: status = %d, want 200", resp.Code)
		}
		var got map[string]int
		if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		want := map[string]int{"count": 5}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GET /stock/apple: body = %v, want %v", got, want)
		}
	})
}

func checkStockList(t *testing.T, mux *http.ServeMux, want []map[string]any) {
	t.Helper()
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/stock", nil))
	if resp.Code != http.StatusOK {
		t.Fatalf("GET /stock: status = %d, want 200", resp.Code)
	}
	if got := resp.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("GET /stock: Content-Type = %q, want application/json", got)
	}
	var got []map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GET /stock: body = %s, want %v", resp.Body.Bytes(), want)
	}
}
