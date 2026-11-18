package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrice(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		status int
		body   string
	}{
		{"fractional price", "unit=0.1&qty=3", http.StatusOK, "{\"price\":0.30000000000000004}\n"},
		{"negative quantity", "unit=2.5&qty=-2", http.StatusOK, "{\"price\":-5}\n"},
		{"negative zero unit", "unit=-0&qty=1", http.StatusOK, "{\"price\":-0}\n"},
		{"negative unit and zero quantity", "unit=-2.5&qty=0", http.StatusOK, "{\"price\":-0}\n"},
		{"negative zero product with positive unit", "unit=0&qty=-1", http.StatusOK, "{\"price\":-0}\n"},
		{"positive zero product", "unit=-0&qty=-1", http.StatusOK, "{\"price\":0}\n"},
		{"invalid unit", "unit=nope&qty=3", http.StatusBadRequest, "bad unit\n"},
		{"invalid quantity", "unit=0.1&qty=nope", http.StatusBadRequest, "bad qty\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handlePrice(w, httptest.NewRequest(http.MethodGet, "/price?"+tt.query, nil))
			if w.Code != tt.status || w.Body.String() != tt.body {
				t.Fatalf("response = (%d, %q), want (%d, %q)", w.Code, w.Body.String(), tt.status, tt.body)
			}
		})
	}
}
