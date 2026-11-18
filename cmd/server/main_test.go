package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockTotal(t *testing.T) {
	server := httptest.NewServer(newMux())
	defer server.Close()
	client := server.Client()

	getQuantity := func(path, key string) int {
		t.Helper()
		resp, err := client.Get(server.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s: status = %d, want 200", path, resp.StatusCode)
		}
		var body map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatalf("GET %s: decode response: %v", path, err)
		}
		qty, ok := body[key]
		if !ok || len(body) != 1 {
			t.Fatalf("GET %s: response = %v, want only %q", path, body, key)
		}
		return qty
	}

	if got := getQuantity("/stock-total", "total"); got != 150 {
		t.Fatalf("seeded stock total = %d, want 150", got)
	}

	order := strings.NewReader(`{"id":"o1","items":[{"name":"apple","unitPrice":2.5,"qty":2}]}`)
	resp, err := client.Post(server.URL+"/orders", "application/json", order)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
	}

	apples := getQuantity("/stock/apple", "count")
	if apples >= 100 {
		t.Fatalf("apple count after order = %d, want less than 100", apples)
	}
	if got, want := getQuantity("/stock-total", "total"), apples+50; got != want {
		t.Fatalf("stock total after order = %d, want %d", got, want)
	}
}
