package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	st.Add("pear", 2)
	server := httptest.NewServer(newMux(st, orders.New(st, &events.Bus{})))
	defer server.Close()

	checkTotal := func(want string) {
		t.Helper()
		resp, err := server.Client().Get(server.URL + "/stock-total")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /stock-total: status = %d, want 200", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("GET /stock-total: Content-Type = %q, want application/json", got)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.TrimSpace(string(body)); got != want {
			t.Errorf("GET /stock-total: body = %q, want %q", got, want)
		}
	}

	checkTotal(`{"total":5}`)

	order := strings.NewReader(`{"id":"1","items":[{"name":"apple","unitPrice":1,"qty":3},{"name":"pear","unitPrice":1,"qty":2}]}`)
	resp, err := server.Client().Post(server.URL+"/orders", "application/json", order)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
	}

	checkTotal(`{"total":0}`)
}

func TestStockTotalRejectsNonGET(t *testing.T) {
	st := stock.NewStock()
	server := httptest.NewServer(newMux(st, orders.New(st, &events.Bus{})))
	defer server.Close()

	for _, method := range []string{http.MethodPost, http.MethodHead} {
		req, err := http.NewRequest(method, server.URL+"/stock-total", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("%s /stock-total: status = %d, want 405", method, resp.StatusCode)
		}
	}
}
