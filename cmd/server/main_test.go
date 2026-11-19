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
	server := httptest.NewServer(mux)
	defer server.Close()

	getTotal := func(t *testing.T, want int) {
		t.Helper()
		res, err := server.Client().Get(server.URL + "/stock-total")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", res.StatusCode, http.StatusOK)
		}
		if got := res.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var body map[string]int
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got, ok := body["total"]; len(body) != 1 || !ok || got != want {
			t.Errorf("body = %v, want only total = %d", body, want)
		}
	}

	t.Run("empty", func(t *testing.T) {
		getTotal(t, 0)
	})
	t.Run("updated", func(t *testing.T) {
		st.Add("apple", 4)
		st.Add("pear", 3)
		getTotal(t, 7)
	})
	t.Run("rejects POST", func(t *testing.T) {
		res, err := server.Client().Post(server.URL+"/stock-total", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want %d", res.StatusCode, http.StatusMethodNotAllowed)
		}
	})
}
