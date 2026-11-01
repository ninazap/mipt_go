package ledger

import (
	"strings"
	"testing"

	"ledger/cache"
	"ledger/db"
)

// setupTestLedger создает экземпляр Ledger с реальной БД для тестов
func setupTestLedger(t *testing.T) *Ledger {
	database, err := db.NewDatabase()
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Кэш опционален для unit-тестов, можно передать nil или реальный клиент
	cacheClient, _ := cache.NewCache()

	l := NewLedger(database, cacheClient)
	if err := l.Reset(); err != nil {
		t.Fatalf("Failed to reset DB: %v", err)
	}
	return l
}

func TestTransactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
		errMsg  string
	}{
		{"valid transaction", Transaction{Amount: 100.0, Category: "Food", Description: "Lunch"}, false, ""},
		{"zero amount", Transaction{Amount: 0, Category: "Food", Description: "Test"}, true, "transaction amount must be greater than zero"},
		{"negative amount", Transaction{Amount: -50.0, Category: "Food", Description: "Test"}, true, "transaction amount must be greater than zero"},
		{"empty category", Transaction{Amount: 100.0, Category: "", Description: "Test"}, true, "transaction category cannot be empty"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestBudgetValidate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
		errMsg  string
	}{
		{"valid budget", Budget{Category: "Food", Limit: 5000.0, Period: "month"}, false, ""},
		{"zero limit", Budget{Category: "Food", Limit: 0, Period: "month"}, true, "budget limit must be positive"},
		{"negative limit", Budget{Category: "Food", Limit: -100.0, Period: "month"}, true, "budget limit must be positive"},
		{"empty category", Budget{Category: "", Limit: 5000.0, Period: "month"}, true, "budget category cannot be empty"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.budget.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestAddTransactionBudgetExceeded(t *testing.T) {
	l := setupTestLedger(t)

	budget := Budget{Category: "Food", Limit: 5000.0, Period: "month"}
	if err := l.SetBudget(budget); err != nil {
		t.Fatalf("SetBudget failed: %v", err)
	}

	tx1 := Transaction{Amount: 3000.0, Category: "Food", Description: "Groceries", Date: "2025-01-01"}
	if err := l.AddTransaction(tx1); err != nil {
		t.Fatalf("AddTransaction tx1 failed: %v", err)
	}

	txs, err := l.ListTransactions()
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	tx2 := Transaction{Amount: 2500.0, Category: "Food", Description: "Expensive dinner", Date: "2025-01-02"}
	err = l.AddTransaction(tx2)
	if err == nil {
		t.Fatal("expected error for budget exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "budget exceeded") {
		t.Errorf("expected 'budget exceeded' error, got: %v", err)
	}

	txs, err = l.ListTransactions()
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction after failed add, got %d", len(txs))
	}
}

func TestSetBudgetValidation(t *testing.T) {
	l := setupTestLedger(t)

	budget := Budget{Category: "", Limit: -100.0}
	err := l.SetBudget(budget)
	if err == nil {
		t.Fatal("expected error for invalid budget, got nil")
	}

	budget = Budget{Category: "Transport", Limit: 2000.0, Period: "month"}
	if err := l.SetBudget(budget); err != nil {
		t.Fatalf("SetBudget failed: %v", err)
	}

	budgets, err := l.ListBudgets()
	if err != nil {
		t.Fatalf("ListBudgets failed: %v", err)
	}
	
	found := false
	for _, b := range budgets {
		if b.Category == "Transport" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Transport' budget to exist")
	}
}
