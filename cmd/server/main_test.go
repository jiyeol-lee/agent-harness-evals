package main

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockTotal(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	defer srv.Close()
	client := srv.Client()

	get := func(t *testing.T, path string) map[string]int {
		t.Helper()
		resp, err := client.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: status %d, want 200", path, resp.StatusCode)
		}
		var body map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		return body
	}

	t.Run("seeded stock", func(t *testing.T) {
		if got := get(t, "/stock-total"); !maps.Equal(got, map[string]int{"total": 150}) {
			t.Errorf("GET /stock-total: got %v, want {total:150}", got)
		}
	})

	t.Run("rejects POST", func(t *testing.T) {
		resp, err := client.Post(srv.URL+"/stock-total", "application/json", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("POST /stock-total: status %d, want 405", resp.StatusCode)
		}
	})

	t.Run("after an order", func(t *testing.T) {
		order := strings.NewReader(`{"id":"order-1","items":[{"name":"apple","unitPrice":2,"qty":3}]}`)
		resp, err := client.Post(srv.URL+"/orders", "application/json", order)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /orders: status %d, want 200", resp.StatusCode)
		}

		want := get(t, "/stock/apple")["count"] + get(t, "/stock/pear")["count"]
		if want >= 150 {
			t.Fatalf("stock after order: got %d, want less than 150", want)
		}
		if got := get(t, "/stock-total"); !maps.Equal(got, map[string]int{"total": want}) {
			t.Errorf("GET /stock-total after order: got %v, want {total:%d}", got, want)
		}
	})
}

func TestStockTotalConcurrentOrder(t *testing.T) {
	srv := httptest.NewServer(newHandler())
	t.Cleanup(srv.Close)
	client := srv.Client()

	t.Run("place order", func(t *testing.T) {
		t.Parallel()
		order := strings.NewReader(`{"id":"order-1","items":[{"name":"apple","unitPrice":2,"qty":3}]}`)
		resp, err := client.Post(srv.URL+"/orders", "application/json", order)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("POST /orders: status %d, want 200", resp.StatusCode)
		}
	})

	t.Run("read total", func(t *testing.T) {
		t.Parallel()
		resp, err := client.Get(srv.URL + "/stock-total")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET /stock-total: status %d, want 200", resp.StatusCode)
		}
	})
}
