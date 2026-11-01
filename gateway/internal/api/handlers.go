package api

import (
	"encoding/json"
	"ledger"
	"net/http"
	"strings"
)

type API struct {
	ledger *ledger.Ledger
}

func NewAPI(ledger *ledger.Ledger) *API {
	return &API{ledger: ledger}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	setJSONHeaders(w)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func (api *API) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx := ToTransaction(req)

	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.ledger.AddTransaction(tx); err != nil {
		if strings.Contains(err.Error(), "budget exceeded") {
			writeError(w, http.StatusConflict, "budget exceeded")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ToTransactionResponse(tx))
}

func (api *API) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Обновлено: теперь обрабатываем ошибку от ListTransactions
	transactions, err := api.ledger.ListTransactions()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get transactions")
		return
	}

	response := ToTransactionsResponse(transactions)
	if response == nil {
		response = []TransactionResponse{}
	}

	writeJSON(w, http.StatusOK, response)
}

func (api *API) CreateBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	budget := ToBudget(req)

	if err := budget.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.ledger.SetBudget(budget); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ToBudgetResponse(budget))
}

func (api *API) ListBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Обновлено: теперь обрабатываем ошибку от ListBudgets
	budgets, err := api.ledger.ListBudgets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get budgets")
		return
	}

	response := ToBudgetsResponse(budgets)
	if response == nil {
		response = []BudgetResponse{}
	}

	writeJSON(w, http.StatusOK, response)
}

func (api *API) GetReportSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to query parameters are required")
		return
	}

	summary, err := api.ledger.GetReportSummary(from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get report summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
