package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestStockList(t *testing.T) {
	tests := []struct {
		name string
		seed bool
		want []map[string]any
	}{
		{name: "empty", want: []map[string]any{}},
		{name: "sorted with zero count", seed: true, want: []map[string]any{
			{"name": "apple", "count": float64(0)},
			{"name": "banana", "count": float64(5)},
			{"name": "pear", "count": float64(2)},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			if tt.seed {
				st.Add("pear", 2)
				st.Add("banana", 3)
				st.Add("apple", 1)
				st.Add("banana", 2)
				if err := st.Remove("apple", 1); err != nil {
					t.Fatal(err)
				}
			}
			mux := http.NewServeMux()
			registerStockRoutes(mux, st)

			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stock", nil))

			if rec.Code != http.StatusOK {
				t.Fatalf("GET /stock status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("GET /stock Content-Type = %q, want application/json", got)
			}
			var got []map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("GET /stock returned invalid JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GET /stock body = %s, want %v", rec.Body, tt.want)
			}
		})
	}
}

func TestStockCount(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 3)
	mux := http.NewServeMux()
	registerStockRoutes(mux, st)

	for _, tt := range []struct {
		name string
		want int
	}{
		{name: "apple", want: 3},
		{name: "unknown", want: 0},
	} {
		path := "/stock/" + tt.name
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want 200", path, rec.Code)
		}
		var got map[string]int
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("GET %s returned invalid JSON: %v", path, err)
		}
		if want := map[string]int{"count": tt.want}; !reflect.DeepEqual(got, want) {
			t.Errorf("GET %s body = %s, want %v", path, rec.Body, want)
		}
	}
}
