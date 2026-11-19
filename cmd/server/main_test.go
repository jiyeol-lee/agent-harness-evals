package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	"example.com/shop/stock"
)

func TestListStock(t *testing.T) {
	populated := stock.NewStock()
	populated.Add("pear", 3)
	populated.Add("banana", 0)
	populated.Add("apple", -2)

	for _, tt := range []struct {
		name string
		st   *stock.Stock
		want []map[string]any
	}{
		{
			name: "sorted rows including zero and negative counts",
			st:   populated,
			want: []map[string]any{
				{"name": "apple", "count": float64(-2)},
				{"name": "banana", "count": float64(0)},
				{"name": "pear", "count": float64(3)},
			},
		},
		{
			name: "empty array",
			st:   stock.NewStock(),
			want: []map[string]any{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/stock", nil)
			res := httptest.NewRecorder()
			listStock(tt.st).ServeHTTP(res, req)

			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			if got := res.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			var got []map[string]any
			if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid response JSON: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("body = %s, want %v", res.Body, tt.want)
			}
		})
	}
}

func TestListStockDuringRemoval(t *testing.T) {
	st := stock.NewStock()
	st.Add("apple", 200)
	h := listStock(st)

	var wg sync.WaitGroup
	wg.Go(func() {
		for qty := 1; qty <= 200; qty++ {
			if err := st.Remove("apple", qty); err != nil {
				t.Error(err)
				return
			}
		}
	})
	for range 200 {
		res := httptest.NewRecorder()
		h.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))
		if res.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", res.Code, http.StatusOK)
		}
	}
	wg.Wait()
}
