package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	tests := []struct {
		name       string
		quantities map[string]int
		want       int
	}{
		{name: "empty inventory", want: 0},
		{name: "multiple items", quantities: map[string]int{"apple": 7, "pear": 3}, want: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := stock.NewStock()
			for item, quantity := range tt.quantities {
				st.Add(item, quantity)
			}

			w := httptest.NewRecorder()
			newServerHandler(st).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/stock-total", nil))

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			var body struct {
				Total *int `json:"total"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body.Total == nil || *body.Total != tt.want {
				t.Errorf("response = %s, want {\"total\":%d}", w.Body, tt.want)
			}
		})
	}
}
