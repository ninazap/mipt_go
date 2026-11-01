package api

import (
	"ledger"
	"net/http"
)

func SetupRouter(ledgerService *ledger.Ledger) http.Handler {
	api := NewAPI(ledgerService)
	
	mux := http.NewServeMux()
	
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

	// Новый маршрут для отчётов
	mux.HandleFunc("/api/reports/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetReportSummary(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
	
	return LoggingMiddleware(mux)
}
