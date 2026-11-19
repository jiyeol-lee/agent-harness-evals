package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestStockTotal(t *testing.T) {
	server := httptest.NewServer(newHandler())
	defer server.Close()
	client := server.Client()

	if got := getStockValue(t, client, server.URL+"/stock-total", "total"); got != 150 {
		t.Fatalf("seeded total = %d, want 150", got)
	}

	resp, err := client.Post(server.URL+"/orders", "application/json", orderBody(3))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /orders: status = %d, want 200", resp.StatusCode)
	}

	apple := getStockValue(t, client, server.URL+"/stock/apple", "count")
	pear := getStockValue(t, client, server.URL+"/stock/pear", "count")
	if apple >= 100 || pear != 50 {
		t.Fatalf("stock after order: apple = %d, pear = %d; want fewer than 100 apples and 50 pears", apple, pear)
	}
	if got := getStockValue(t, client, server.URL+"/stock-total", "total"); got != apple+pear {
		t.Fatalf("total after order = %d, want %d", got, apple+pear)
	}
}

func TestStockRoutesConcurrent(t *testing.T) {
	handler := newHandler()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, route := range []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/orders"},
		{http.MethodGet, "/stock-total"},
		{http.MethodGet, "/stock/apple"},
	} {
		wg.Go(func() {
			<-start
			for i := 1; i <= 10; i++ {
				var req *http.Request
				if route.method == http.MethodPost {
					req = httptest.NewRequest(route.method, route.path, orderBody(i))
				} else {
					req = httptest.NewRequest(route.method, route.path, nil)
				}
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					t.Errorf("%s %s: status = %d, want 200", route.method, route.path, rec.Code)
					return
				}
			}
		})
	}
	close(start)
	wg.Wait()
}

func getStockValue(t *testing.T, client *http.Client, url, key string) int {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status = %d, want 200", url, resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); key == "total" && got != "application/json" {
		t.Fatalf("GET %s: Content-Type = %q, want application/json", url, got)
	}
	var body map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("GET %s: decode body: %v", url, err)
	}
	value, ok := body[key]
	if !ok {
		t.Fatalf("GET %s: missing %q in %v", url, key, body)
	}
	return value
}

func orderBody(qty int) *strings.Reader {
	return strings.NewReader(fmt.Sprintf(`{"id":"order-%d","items":[{"name":"apple","unitPrice":2.5,"qty":%d}]}`, qty, qty))
}
