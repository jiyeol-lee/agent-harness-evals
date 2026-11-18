package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestGetStock(t *testing.T) {
	st := stock.NewStock()
	server := httptest.NewServer(routes(st))
	defer server.Close()

	check := func(want []map[string]any) {
		t.Helper()
		resp, err := server.Client().Get(server.URL + "/stock")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected response: %s, %v", resp.Status, resp.Header)
		}
		var got []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GET /stock = %#v, want %#v", got, want)
		}
	}
	check([]map[string]any{})
	st.Add("pear", 0)
	st.Add("apple", 12)
	check([]map[string]any{{"name": "apple", "count": 12.0}, {"name": "pear", "count": 0.0}})
}
