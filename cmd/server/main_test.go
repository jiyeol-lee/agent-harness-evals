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
	server := httptest.NewServer(stockTotalHandler(st))
	defer server.Close()

	check := func(want int) {
		t.Helper()
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		var got map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		total, ok := got["total"]
		if len(got) != 1 || !ok || total != want {
			t.Fatalf("response = %v, want total %d", got, want)
		}
	}
	check(0)
	st.Add("apple", 3)
	st.Add("pear", 4)
	check(7)
	if err := st.Remove("apple", 2); err != nil {
		t.Fatal(err)
	}
	check(5)
}
