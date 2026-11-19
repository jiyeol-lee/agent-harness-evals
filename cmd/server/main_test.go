package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestStockList(t *testing.T) {
	st := stock.NewStock()
	mux := http.NewServeMux()
	registerStockList(mux, st)

	check := func(t *testing.T, want []map[string]any) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var got []map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("invalid stock list: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("stock list = %#v, want %#v", got, want)
		}
	}

	t.Run("empty", func(t *testing.T) {
		check(t, []map[string]any{})
	})

	t.Run("sorted current counts", func(t *testing.T) {
		st.Add("pear", 2)
		st.Add("banana", 1)
		st.Add("apple", 3)
		st.Add("pear", 4)
		if err := st.Remove("banana", 2); err != nil {
			t.Fatal(err)
		}
		if err := st.Remove("apple", 3); err != nil {
			t.Fatal(err)
		}

		check(t, []map[string]any{
			{"name": "apple", "count": float64(0)},
			{"name": "banana", "count": float64(-1)},
			{"name": "pear", "count": float64(6)},
		})
	})
}
