package main

import (
	"encoding/json"
	"net/http"

	"example.com/shop/stock"
)

func registerStockTotal(mux *http.ServeMux, st *stock.Stock) {
	mux.HandleFunc("GET /stock-total", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total": st.Total()})
	})
}
