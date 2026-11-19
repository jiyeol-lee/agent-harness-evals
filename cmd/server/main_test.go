package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestGetStockTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	st.Add("pear", 5)

	mux := http.NewServeMux()
	registerStockRoutes(mux, st)
	server := httptest.NewServer(mux)
	defer server.Close()

	checkTotal := func(want int) {
		t.Helper()
		resp, err := server.Client().Get(server.URL + "/stock-total")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /stock-total status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("GET /stock-total Content-Type = %q, want application/json", got)
		}

		var body map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if got, ok := body["total"]; !ok || got != want || len(body) != 1 {
			t.Fatalf("GET /stock-total body = %v, want {\"total\":%d}", body, want)
		}
	}

	checkTotal(8)
	st.Add("apple", 2)
	if err := st.Remove("pear", 1); err != nil {
		t.Fatal(err)
	}
	checkTotal(9)
}
