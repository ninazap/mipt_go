package api

import (
	"ledger"
	"net/http"
)

// SetupRouter настраивает маршруты API
func SetupRouter(ledger *ledger.Ledger) http.Handler {
	api := NewAPI(ledger)
	
	mux := http.NewServeMux()
	
	// Маршруты для транзакций
	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.ListTransactions(w, r)
		case http.MethodPost:
			api.CreateTransaction(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
	
	// Маршруты для бюджетов
	mux.HandleFunc("/api/budgets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.ListBudgets(w, r)
		case http.MethodPost:
			api.CreateBudget(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
	
	// Применяем middleware
	return LoggingMiddleware(mux)
}