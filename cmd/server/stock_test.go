package main

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"example.com/shop/stock"
)

func TestGetStock(t *testing.T) {
	st := stock.NewStock()
	handler := stockListHandler(st)
	type item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	check := func(want []item) {
		t.Helper()
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest("GET", "/stock", nil))
		var got []item
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if rec.Code != 200 || !reflect.DeepEqual(got, want) {
			t.Fatalf("GET /stock returned %d %s, want 200 %#v", rec.Code, rec.Body.String(), want)
		}
	}
	check([]item{})
	st.Add("pear", 4)
	st.Add("apple", 1)
	st.Add("pear", 2)
	if err := st.Remove("apple", 1); err != nil {
		t.Fatal(err)
	}
	check([]item{{"apple", 0}, {"pear", 6}})
}
