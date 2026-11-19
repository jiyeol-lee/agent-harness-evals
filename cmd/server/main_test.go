package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/shop/stock"
)

func TestGetStock(t *testing.T) {
	st := stock.NewStock()
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	t.Run("empty", func(t *testing.T) {
		assertStockResponse(t, server, "/stock", `[]`)
	})

	t.Run("current counts sorted by name", func(t *testing.T) {
		st.Add("pear", 4)
		st.Add("apple", 2)
		st.Add("pear", 1)
		if err := st.Remove("apple", 2); err != nil {
			t.Fatal(err)
		}

		assertStockResponse(t, server, "/stock", `[{"name":"apple","count":0},{"name":"pear","count":5}]`)
		assertStockResponse(t, server, "/stock/pear", `{"count":5}`)
	})
}

func assertStockResponse(t *testing.T, server *httptest.Server, path, want string) {
	t.Helper()
	res, err := server.Client().Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", path, res.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(body)); got != want {
		t.Fatalf("GET %s body = %s, want %s", path, got, want)
	}
	if path == "/stock" {
		if got := res.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("GET %s Content-Type = %q, want application/json", path, got)
		}
	}
}
