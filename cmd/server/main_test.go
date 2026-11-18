package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPriceZeroSign(t *testing.T) {
	mux := http.NewServeMux()
	registerPrice(mux)
	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"negative unit", "unit=-0&qty=1", `{"price":-0}` + "\n"},
		{"negative quantity", "unit=0&qty=-1", `{"price":-0}` + "\n"},
		{"zero quantity with negative unit", "unit=-5&qty=0", `{"price":-0}` + "\n"},
		{"zero quantity with positive unit", "unit=5&qty=0", `{"price":0}` + "\n"},
		{"two negatives", "unit=-0&qty=-1", `{"price":0}` + "\n"},
		{"two positives", "unit=0&qty=1", `{"price":0}` + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := server.Client().Get(server.URL + "/price?" + tt.query)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK || string(body) != tt.want {
				t.Fatalf("GET /price?%s = %d %q, want 200 %q", tt.query, resp.StatusCode, body, tt.want)
			}
		})
	}
}
