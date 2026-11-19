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
	for _, tt := range []struct {
		name string
		path string
		seed bool
		want string
	}{
		{"list", "/stock", true, `[{"name":"apple","count":2},{"name":"pear","count":0}]`},
		{"empty list", "/stock", false, `[]`},
		{"lookup", "/stock/apple", true, `{"count":2}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			if tt.seed {
				st.Add("pear", 1)
				st.Add("apple", 2)
				if err := st.Remove("pear", 1); err != nil {
					t.Fatal(err)
				}
			}
			mux := http.NewServeMux()
			registerStockRoutes(mux, st)

			res := httptest.NewRecorder()
			mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if res.Code != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want %d", tt.path, res.Code, http.StatusOK)
			}
			if tt.path == "/stock" {
				if got := res.Header().Get("Content-Type"); got != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", got)
				}
			}
			var got, want any
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatalf("invalid expected JSON: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body = %s, want %s", res.Body.Bytes(), tt.want)
			}
		})
	}
}
