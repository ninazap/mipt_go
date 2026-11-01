package api

import "ledger"

// DTO для запросов и ответов

type CreateTransactionRequest struct {
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type TransactionResponse struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
}

type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period,omitempty"`
}

type BudgetResponse struct {
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
}

// Преобразователи DTO <-> Domain models

func ToTransaction(req CreateTransactionRequest) ledger.Transaction {
	return ledger.Transaction{
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
	}
}

func ToTransactionResponse(tx ledger.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		Amount:      tx.Amount,
		Category:    tx.Category,
		Description: tx.Description,
		Date:        tx.Date,
	}
}

func ToTransactionsResponse(txs []ledger.Transaction) []TransactionResponse {
	result := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		result = append(result, ToTransactionResponse(tx))
	}
	return result
}

func ToBudget(req CreateBudgetRequest) ledger.Budget {
	return ledger.Budget{
		Category: req.Category,
		Limit:    req.Limit,
		Period:   req.Period,
	}
}

func ToBudgetResponse(b ledger.Budget) BudgetResponse {
	return BudgetResponse{
		Category: b.Category,
		Limit:    b.Limit,
		Period:   b.Period,
	}
}

func ToBudgetsResponse(budgets map[string]ledger.Budget) []BudgetResponse {
	result := make([]BudgetResponse, 0, len(budgets))
	for _, b := range budgets {
		result = append(result, ToBudgetResponse(b))
	}
	return result
}