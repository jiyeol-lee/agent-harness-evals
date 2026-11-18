// Command server serves the order API.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"example.com/shop/events"
	"example.com/shop/orders"
	"example.com/shop/pricing"
	"example.com/shop/stock"
)

func main() {
	log.Fatal(http.ListenAndServe(":8080", newHandler()))
}

func newHandler() http.Handler {
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

	mux.HandleFunc("GET /stock/{name}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]int{"count": st.Count(r.PathValue("name"))})
	})

	mux.HandleFunc("GET /stock-total", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total": st.Total()})
	})

	return mux
}
