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

// ErrorResponse структура для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

// setJSONHeaders устанавливает заголовки для JSON ответа
func setJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// writeJSON записывает JSON ответ
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	setJSONHeaders(w)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError записывает ошибку в формате JSON
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// CreateTransaction обработчик POST /api/transactions
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

	// Валидация
	if err := tx.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Добавление транзакции
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

// ListTransactions обработчик GET /api/transactions
func (api *API) ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	transactions := api.ledger.ListTransactions()
	response := ToTransactionsResponse(transactions)
	
	// Если транзакций нет, возвращаем пустой массив
	if response == nil {
		response = []TransactionResponse{}
	}
	
	writeJSON(w, http.StatusOK, response)
}

// CreateBudget обработчик POST /api/budgets
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

	// Валидация
	if err := budget.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Добавление/обновление бюджета
	if err := api.ledger.SetBudget(budget); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, ToBudgetResponse(budget))
}

// ListBudgets обработчик GET /api/budgets
func (api *API) ListBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	budgets := api.ledger.ListBudgets()
	response := ToBudgetsResponse(budgets)
	
	// Если бюджетов нет, возвращаем пустой массив
	if response == nil {
		response = []BudgetResponse{}
	}
	
	writeJSON(w, http.StatusOK, response)
}