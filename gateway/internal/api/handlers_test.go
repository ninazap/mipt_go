package api

import (
	"encoding/json"
	"ledger"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setupTestAPI создаёт тестовый API с изолированным Ledger
func setupTestAPI() (*API, *ledger.Ledger) {
	l := ledger.NewLedger()
	api := NewAPI(l)
	return api, l
}

// TestCreateBudget_Valid — валидный POST /api/budgets
func TestCreateBudget_Valid(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	body := `{"category":"Food","limit":5000,"period":"month"}`
	req := httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.CreateBudget(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type 'application/json; charset=utf-8', got %q", contentType)
	}

	var budgetResp BudgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&budgetResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if budgetResp.Category != "Food" || budgetResp.Limit != 5000 {
		t.Errorf("unexpected budget response: %+v", budgetResp)
	}
}

// TestCreateBudget_Invalid — невалидный POST /api/budgets
func TestCreateBudget_Invalid(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "zero limit",
			body:       `{"category":"Food","limit":0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty category",
			body:       `{"category":"","limit":5000}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			api.CreateBudget(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}
			if errResp.Error == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}

// TestListBudgets — GET /api/budgets
func TestListBudgets(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	// Добавляем бюджет
	l.SetBudget(ledger.Budget{Category: "Food", Limit: 5000.0, Period: "month"})

	req := httptest.NewRequest(http.MethodGet, "/api/budgets", nil)
	w := httptest.NewRecorder()

	api.ListBudgets(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var budgets []BudgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&budgets); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(budgets) != 1 {
		t.Errorf("expected 1 budget, got %d", len(budgets))
	}
	if budgets[0].Category != "Food" {
		t.Errorf("expected category 'Food', got %q", budgets[0].Category)
	}
}

// TestCreateTransaction_Valid — валидный POST /api/transactions
func TestCreateTransaction_Valid(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	// Устанавливаем бюджет
	l.SetBudget(ledger.Budget{Category: "Food", Limit: 5000.0, Period: "month"})

	body := `{"amount":1000,"category":"Food","description":"Lunch","date":"2025-09-10"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var txResp TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&txResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if txResp.Amount != 1000 || txResp.Category != "Food" {
		t.Errorf("unexpected transaction response: %+v", txResp)
	}
}

// TestListTransactions — GET /api/transactions
func TestListTransactions(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	// Добавляем транзакцию
	l.AddTransaction(ledger.Transaction{Amount: 1000.0, Category: "Food", Description: "Lunch"})

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()

	api.ListTransactions(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var txs []TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

// TestCreateTransaction_EmptyList — GET /api/transactions возвращает пустой массив
func TestCreateTransaction_EmptyList(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()

	api.ListTransactions(w, req)

	resp := w.Result()
	var txs []TransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Проверяем, что это пустой массив, а не null
	body := strings.TrimSpace(w.Body.String())
	if body != "[]" {
		t.Errorf("expected empty array '[]', got %q", body)
	}
}

// TestCreateTransaction_BudgetExceeded — превышение бюджета
func TestCreateTransaction_BudgetExceeded(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	// Устанавливаем бюджет
	l.SetBudget(ledger.Budget{Category: "Food", Limit: 5000.0, Period: "month"})

	// Добавляем транзакцию в пределах лимита
	body1 := `{"amount":4500,"category":"Food","description":"Groceries","date":"2025-09-10"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	api.CreateTransaction(w1, req1)

	if w1.Result().StatusCode != http.StatusCreated {
		t.Fatalf("first transaction should succeed, got status %d", w1.Result().StatusCode)
	}

	// Пытаемся добавить транзакцию, которая превысит лимит
	body2 := `{"amount":1000,"category":"Food","description":"Dinner","date":"2025-09-10"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	api.CreateTransaction(w2, req2)

	resp := w2.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error != "budget exceeded" {
		t.Errorf("expected error 'budget exceeded', got %q", errResp.Error)
	}

	// Проверяем, что транзакция не добавлена
	txs := l.ListTransactions()
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

// TestCreateTransaction_InvalidJSON — невалидный JSON
func TestCreateTransaction_InvalidJSON(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

// TestCreateTransaction_ValidationError — ошибка валидации
func TestCreateTransaction_ValidationError(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	body := `{"amount":-100,"category":"Food","description":"Invalid","date":"2025-09-10"}`
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.CreateTransaction(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// TestIntegration_FullWorkflow — интеграционный тест полного сценария
func TestIntegration_FullWorkflow(t *testing.T) {
	api, l := setupTestAPI()
	defer l.Reset()

	t.Run("create budget", func(t *testing.T) {
		body := `{"category":"Food","limit":5000,"period":"month"}`
		req := httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.CreateBudget(w, req)

		if w.Result().StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Result().StatusCode)
		}
	})

	t.Run("create valid transaction", func(t *testing.T) {
		body := `{"amount":1000,"category":"Food","description":"Lunch","date":"2025-09-10"}`
		req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.CreateTransaction(w, req)

		if w.Result().StatusCode != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Result().StatusCode)
		}
	})

	t.Run("list transactions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
		w := httptest.NewRecorder()

		api.ListTransactions(w, req)

		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Result().StatusCode)
		}

		var txs []TransactionResponse
		if err := json.NewDecoder(w.Result().Body).Decode(&txs); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if len(txs) != 1 {
			t.Errorf("expected 1 transaction, got %d", len(txs))
		}
	})

	t.Run("exceed budget", func(t *testing.T) {
		body := `{"amount":4500,"category":"Food","description":"Dinner","date":"2025-09-10"}`
		req := httptest.NewRequest(http.MethodPost, "/api/transactions", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		api.CreateTransaction(w, req)

		if w.Result().StatusCode != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Result().StatusCode)
		}
	})
}

// TestMiddleware_LoggingMiddleware — тест middleware (базовый)
func TestMiddleware_LoggingMiddleware(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Result().StatusCode)
	}
}