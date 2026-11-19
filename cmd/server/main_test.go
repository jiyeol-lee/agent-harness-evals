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

		body := strings.NewReader(`{"id":"one","items":[{"name":"apple","unitPrice":1,"qty":1}]}`)
		resp, err := server.Client().Post(server.URL+"/orders", "application/json", body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
		}

		got := stockTotal(t, server)
		if got >= 150 {
			t.Errorf("stock total after order = %d, want less than 150", got)
		}
		var item struct {
			Count int `json:"count"`
		}
		getStockJSON(t, server, "/stock/apple", &item)
		if want := item.Count + 50; got != want {
			t.Errorf("stock total after order = %d, want %d (apple count plus 50 pears)", got, want)
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
	var body struct {
		Total *int `json:"total"`
	}
	getStockJSON(t, server, "/stock-total", &body)
	if body.Total == nil {
		t.Fatal("GET /stock-total: missing total")
	}
	return *body.Total
}

func getStockJSON(t *testing.T, server *httptest.Server, path string, dst any) {
	t.Helper()
	resp, err := server.Client().Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want 200", path, resp.StatusCode)
	}
	if path == "/stock-total" {
		if got := resp.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("GET %s: Content-Type = %q, want application/json", path, got)
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("GET %s: decode JSON: %v", path, err)
	}
}
