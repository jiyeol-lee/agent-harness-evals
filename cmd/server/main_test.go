package main

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()

	get := func(path, key string) int {
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
			t.Fatalf("GET %s: decode JSON: %v", path, err)
		}
		value, ok := body[key]
		if !ok || len(body) != 1 {
			t.Fatalf("GET %s: body = %v, want only %q", path, body, key)
		}
		return value
	}

	if got, want := get("/stock-total", "total"), 150; got != want {
		t.Fatalf("seeded stock total = %v, want %v", got, want)
	}

	order := strings.NewReader(`{"id":"o1","items":[{"name":"apple","unitPrice":2,"qty":3}]}`)
	resp, err := client.Post(server.URL+"/orders", "application/json", order)
	if err != nil {
		t.Fatal(err)
	}
	_, readErr := io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
	}

	pear := get("/stock/pear", "count")
	if want := 50; pear != want {
		t.Fatalf("pear stock = %v, want %v", pear, want)
	}
	apple := get("/stock/apple", "count")
	if apple >= 100 {
		t.Fatalf("apple stock after order = %d, want less than 100", apple)
	}
	if got, want := get("/stock-total", "total"), apple+pear; got != want {
		t.Fatalf("updated stock total = %v, want %v", got, want)
	}
}

func TestStockTotalBeyondMaxInt(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", math.MaxInt)
	st.Add("pear", 2)
	server := httptest.NewServer(routes(st))
	defer server.Close()

	resp, err := server.Client().Get(server.URL + "/stock-total")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /stock-total: status = %d, want 200", resp.StatusCode)
	}
	var body map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("GET /stock-total: decode JSON: %v", err)
	}
	want := strconv.FormatUint(uint64(math.MaxInt)+2, 10)
	if got := string(body["total"]); got != want || len(body) != 1 {
		t.Fatalf("stock total = %s, want unquoted %s", got, want)
	}
}
