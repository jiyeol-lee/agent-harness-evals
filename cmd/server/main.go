// Command server serves the order API.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/pricing"
	"example.com/shop/stock"
)

func main() {
	st := stock.NewStock()
	st.Add("apple", 100)
	st.Add("pear", 50)

	bus := &events.Bus{}
	bus.Subscribe(func(e events.OrderPlaced) {
		for name, qty := range e.Items {
			_ = st.Remove(name, qty)
		}
	})
	svc := orders.New(st, bus)

	http.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID    string        `json:"id"`
			Items []orders.Item `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		total, err := svc.Place(req.ID, req.Items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		json.NewEncoder(w).Encode(map[string]float64{"total": total})
	})

	http.HandleFunc("GET /price", func(w http.ResponseWriter, r *http.Request) {
		unit, err := strconv.ParseFloat(r.URL.Query().Get("unit"), 64)
		if err != nil {
			http.Error(w, "bad unit", http.StatusBadRequest)
			return
		}
		qty, err := strconv.Atoi(r.URL.Query().Get("qty"))
		if err != nil {
			http.Error(w, "bad qty", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]float64{"price": pricing.CalcPrice(unit, qty)})
	})

	registerStockRoutes(http.DefaultServeMux, st)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerStockRoutes(mux *http.ServeMux, st *stock.Stock) {
	mux.HandleFunc("GET /stock", func(w http.ResponseWriter, r *http.Request) {
		type stockRow struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		counts := st.Counts()
		rows := make([]stockRow, 0, len(counts))
		for name, count := range counts {
			rows = append(rows, stockRow{Name: name, Count: count})
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Name < rows[j].Name
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	})

	mux.HandleFunc("GET /stock/{name}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]int{"count": st.Count(r.PathValue("name"))})
	})
}
