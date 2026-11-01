package api

import (
	"encoding/json"
	"ledger"
	"ledger/cache"
	"ledger/db"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setupTestAPI(t *testing.T) (*API, *ledger.Ledger) {
	// Если переменная DATABASE_URL не задана, пропускаем интеграционные тесты
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("Skipping test: DATABASE_URL is not set")
	}

	database, err := db.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	
	cacheClient, _ := cache.NewCache() // Кеш опционален для тестов

	l := ledger.NewLedger(database, cacheClient)
	
	// Очищаем БД перед тестом
	if err := l.Reset(); err != nil {
		t.Fatalf("Failed to reset DB: %v", err)
	}

	return NewAPI(l), l
}

func TestCreateBudget_Valid(t *testing.T) {
	api, l := setupTestAPI(t)
	defer l.Reset()

	body := `{"category":"Food","limit":5000,"period":"month"}`
	req := httptest.NewRequest(http.MethodPost, "/api/budgets", strings.NewReader(body))
	req.Header.Set("Content-Type", "apkication/json")
	w := httptest.NewRecorder()

	api.CreateBudget(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Result().StatusCode)
	}
}

func TestListTransactions(t *testing.T) {
	api, l := setupTestAPI(t)
	defer l.Reset()

	// Добавляем тестовую транзакцию напрямую
	l.AddTransaction(ledger.Transaction{Amount: 1000.0, Category: "Food", Description: "Test", Date: "2025-01-01"})

	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()

	api.ListTransactions(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Result().StatusCode)
	}

	var txs []map[string]interface{}
	if err := json.NewDecoder(w.Result().Body).Decode(&txs); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}
}

func TestGetReportSummary(t *testing.T) {
	api, l := setupTestAPI(t)
	defer l.Reset()

	l.AddTransaction(ledger.Transaction{Amount: 500.0, Category: "Food", Description: "Test", Date: "2025-09-15"})

	req := httptest.NewRequest(http.MethodGet, "/api/reports/summary?from=2025-09-01&to=2025-09-30", nil)
	w := httptest.NewRecorder()

	api.GetReportSummary(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Result().StatusCode)
	}
}
