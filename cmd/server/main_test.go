package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/stock"
)

func TestGetStock(t *testing.T) {
	type item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	st := stock.NewStock()
	st.Add("pear", 2)
	st.Add("banana", 0)
	st.Add("apple", 3)
	st.Add("pear", 4)
	if err := st.Remove("apple", 1); err != nil {
		t.Fatal(err)
	}

	mux := newMux(st, orders.New(st, &events.Bus{}))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("GET /stock status = %d, want 200", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got []item
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := []item{{Name: "apple", Count: 2}, {Name: "banana", Count: 0}, {Name: "pear", Count: 6}}
	if !slices.Equal(got, want) {
		t.Errorf("GET /stock = %+v, want %+v", got, want)
	}
}

func TestGetStockEmpty(t *testing.T) {
	st := stock.NewStock()
	mux := newMux(st, orders.New(st, &events.Bus{}))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/stock", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("GET /stock status = %d, want 200", res.Code)
	}
	var got []json.RawMessage
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("GET /stock = %s, want []", res.Body.Bytes())
	}
}
