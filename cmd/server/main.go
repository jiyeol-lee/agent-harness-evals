// Command server serves the order API.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

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

	log.Fatal(http.ListenAndServe(":8080", newMux(st, svc)))
}

func newMux(st *stock.Stock, svc *orders.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("GET /price", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("GET /stock", func(w http.ResponseWriter, r *http.Request) {
		items := st.Snapshot()
		slices.SortFunc(items, func(a, b stock.ItemCount) int {
			return strings.Compare(a.Name, b.Name)
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	})

	mux.HandleFunc("GET /stock/{name}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]int{"count": st.Count(r.PathValue("name"))})
	})

	return mux
}
