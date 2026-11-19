package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPriceHandler(t *testing.T) {
	tests := []struct {
		query  string
		status int
		body   string
	}{
		{"unit=0.1&qty=3", http.StatusOK, "{\"price\":0.30000000000000004}\n"},
		{"unit=-2.5&qty=0", http.StatusOK, "{\"price\":-0}\n"},
		{"unit=-0&qty=2", http.StatusOK, "{\"price\":-0}\n"},
		{"unit=0&qty=-2", http.StatusOK, "{\"price\":-0}\n"},
		{"unit=-0&qty=-2", http.StatusOK, "{\"price\":0}\n"},
		{"unit=bad&qty=bad", http.StatusBadRequest, "bad unit\n"},
		{"unit=2.5&qty=bad", http.StatusBadRequest, "bad qty\n"},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			w := httptest.NewRecorder()
			priceHandler(w, httptest.NewRequest(http.MethodGet, "/price?"+tt.query, nil))
			if w.Code != tt.status || w.Body.String() != tt.body {
				t.Errorf("response = (%d, %q), want (%d, %q)", w.Code, w.Body.String(), tt.status, tt.body)
			}
		})
	}
}
