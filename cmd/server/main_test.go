package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		server := httptest.NewServer(newHandler(stock.NewStock()))
		defer server.Close()

		if got := stockTotal(t, server); got != 0 {
			t.Errorf("empty stock total = %d, want 0", got)
		}
	})

	t.Run("live", func(t *testing.T) {
		st := stock.NewStock()
		st.Add("apple", 100)
		st.Add("pear", 50)
		server := httptest.NewServer(newHandler(st))
		defer server.Close()

		if got := stockTotal(t, server); got != 150 {
			t.Fatalf("initial stock total = %d, want 150", got)
		}

		body := strings.NewReader(`{"id":"one","items":[{"name":"apple","unitPrice":1,"qty":2},{"name":"apple","unitPrice":1,"qty":3}]}`)
		resp, err := server.Client().Post(server.URL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
		}

		// The reservation removes 2 and 3; the existing subscriber removes 5.
		// With distinct consecutive quantities, stock's time-based dedupe never applies.
		if got := stockTotal(t, server); got != 140 {
			t.Errorf("stock total after order = %d, want 140", got)
		}
	})
}

func TestStockConcurrentRequests(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 1000)
	server := httptest.NewServer(newHandler(st))
	defer server.Close()

	start := make(chan struct{})
	var requests sync.WaitGroup
	for i := 1; i <= 12; i++ {
		for _, path := range []string{"/orders", "/stock-total", "/stock/apple"} {
			requests.Go(func() {
				<-start
				var resp *http.Response
				var err error
				if path == "/orders" {
					body := fmt.Sprintf(`{"id":"%d","items":[{"name":"apple","unitPrice":1,"qty":%d}]}`, i, i)
					resp, err = server.Client().Post(server.URL+path, "application/json", strings.NewReader(body))
				} else {
					resp, err = server.Client().Get(server.URL + path)
				}
				if err != nil {
					t.Errorf("%s: %v", path, err)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("%s: status = %d, want 200", path, resp.StatusCode)
				}
			})
		}
	}
	close(start)
	requests.Wait()
}

func stockTotal(t *testing.T, server *httptest.Server) int {
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
	var body struct {
		Total *int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("GET /stock-total: decode JSON: %v", err)
	}
	if body.Total == nil {
		t.Fatal("GET /stock-total: missing total")
	}
	return *body.Total
}
