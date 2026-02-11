package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/WaiperOK/llm-gateway-control-plane/internal/router"
)

func main() {
	http.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method_not_allowed", http.StatusMethodNotAllowed)
			return
		}
		var req router.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid_json", http.StatusBadRequest)
			return
		}
		budget := router.TenantBudget{MonthlyTokenLimit: 500_000, UsedTokens: 120_000}
		decision := router.Decide(req, budget)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(decision)
	})

	log.Println("llm-gateway-control-plane listening on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
