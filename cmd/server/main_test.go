package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStockRoutes(t *testing.T) {
	handler := newHandler()
	check := func(method, path, body, want string) {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != want {
			t.Fatalf("%s %s: status %d, body %q; want 200, %q", method, path, rec.Code, rec.Body.String(), want)
		}
	}
	check("GET", "/stock", "", `[{"name":"apple","count":100},{"name":"pear","count":50}]`)
	check("GET", "/stock/apple", "", `{"count":100}`)
	check("POST", "/orders", `{"id":"1","items":[{"Name":"apple","UnitPrice":1,"Qty":2}]}`, `{"total":2}`)
	check("GET", "/stock", "", `[{"name":"apple","count":98},{"name":"pear","count":50}]`)
	check("GET", "/stock/apple", "", `{"count":98}`)
}
