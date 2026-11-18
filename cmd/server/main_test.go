package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"example.com/shop/stock"
)

func TestStockTotal(t *testing.T) {
	st := stock.NewStock()
	server := httptest.NewServer(stockTotalHandler(st))
	defer server.Close()

	check := func(want int) {
		t.Helper()
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var got map[string]int
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		total, ok := got["total"]
		if len(got) != 1 || !ok || total != want {
			t.Fatalf("response = %v, want total %d", got, want)
		}
	}
	check(0)
	st.Add("apple", 3)
	st.Add("pear", 4)
	check(7)
	if err := st.Remove("apple", 2); err != nil {
		t.Fatal(err)
	}
	check(5)
}

func TestStockTotalBeyondInt(t *testing.T) {
	st := stock.NewStock()
	maxInt := int(^uint(0) >> 1)
	st.Add("apple", maxInt)
	st.Add("pear", maxInt)
	st.Add("plum", 1)
	server := httptest.NewServer(stockTotalHandler(st))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"total\":18446744073709551615}\n"
	if strconv.IntSize == 32 {
		want = "{\"total\":4294967295}\n"
	}
	if string(body) != want {
		t.Fatalf("response = %q, want %q", body, want)
	}
}
