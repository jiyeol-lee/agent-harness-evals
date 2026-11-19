package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPrice(t *testing.T) {
	mux := http.NewServeMux()
	registerPrice(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	tests := []struct {
		name   string
		query  string
		status int
		body   string
	}{
		{"positive", "unit=2.5&qty=3", http.StatusOK, "{\"price\":7.5}\n"},
		{"negative quantity", "unit=2.5&qty=-3", http.StatusOK, "{\"price\":-7.5}\n"},
		{"negative zero unit", "unit=-0&qty=2", http.StatusOK, "{\"price\":-0}\n"},
		{"negative unit with zero quantity", "unit=-2&qty=0", http.StatusOK, "{\"price\":-0}\n"},
		{"positive zero unit with negative quantity", "unit=0&qty=-2", http.StatusOK, "{\"price\":-0}\n"},
		{"negative zero unit with negative quantity", "unit=-0&qty=-2", http.StatusOK, "{\"price\":0}\n"},
		{"invalid unit first", "unit=bad&qty=bad", http.StatusBadRequest, "bad unit\n"},
		{"invalid quantity", "unit=2.5&qty=bad", http.StatusBadRequest, "bad qty\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := srv.Client().Get(srv.URL + "/price?" + tt.query)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status || string(body) != tt.body {
				t.Fatalf("response = (%d, %q), want (%d, %q)", resp.StatusCode, body, tt.status, tt.body)
			}
		})
	}
}
