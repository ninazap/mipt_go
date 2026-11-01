package ledger

import (
	"strings"
	"testing"
)

// TestTransactionValidate — table-driven тесты для Transaction.Validate()
func TestTransactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		tx      Transaction
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid transaction",
			tx:      Transaction{Amount: 100.0, Category: "Food", Description: "Lunch"},
			wantErr: false,
		},
		{
			name:    "zero amount",
			tx:      Transaction{Amount: 0, Category: "Food", Description: "Test"},
			wantErr: true,
			errMsg:  "transaction amount must be greater than zero",
		},
		{
			name:    "negative amount",
			tx:      Transaction{Amount: -50.0, Category: "Food", Description: "Test"},
			wantErr: true,
			errMsg:  "transaction amount must be greater than zero",
		},
		{
			name:    "empty category",
			tx:      Transaction{Amount: 100.0, Category: "", Description: "Test"},
			wantErr: true,
			errMsg:  "transaction category cannot be empty",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
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

// TestBudgetValidate — table-driven тесты для Budget.Validate()
func TestBudgetValidate(t *testing.T) {
	tests := []struct {
		name    string
		budget  Budget
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid budget",
			budget:  Budget{Category: "Food", Limit: 5000.0, Period: "month"},
			wantErr: false,
		},
		{
			name:    "zero limit",
			budget:  Budget{Category: "Food", Limit: 0, Period: "month"},
			wantErr: true,
			errMsg:  "budget limit must be positive",
		},
		{
			name:    "negative limit",
			budget:  Budget{Category: "Food", Limit: -100.0, Period: "month"},
			wantErr: true,
			errMsg:  "budget limit must be positive",
		},
		{
			name:    "empty category",
			budget:  Budget{Category: "", Limit: 5000.0, Period: "month"},
			wantErr: true,
			errMsg:  "budget category cannot be empty",
		},
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

// TestAddTransactionBudgetExceeded — тест бизнес-правила превышения бюджета
func TestAddTransactionBudgetExceeded(t *testing.T) {
	l := NewLedger()
	defer l.Reset()

	// Устанавливаем бюджет
	budget := Budget{Category: "Food", Limit: 5000.0, Period: "month"}
	if err := l.SetBudget(budget); err != nil {
		t.Fatalf("SetBudget failed: %v", err)
	}

	// Добавляем транзакцию в пределах лимита
	tx1 := Transaction{Amount: 3000.0, Category: "Food", Description: "Groceries"}
	if err := l.AddTransaction(tx1); err != nil {
		t.Fatalf("AddTransaction tx1 failed: %v", err)
	}

	// Проверяем, что транзакция добавлена
	txs := l.ListTransactions()
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	// Пытаемся добавить транзакцию, которая превысит лимит
	tx2 := Transaction{Amount: 2500.0, Category: "Food", Description: "Expensive dinner"}
	err := l.AddTransaction(tx2)
	if err == nil {
		t.Fatal("expected error for budget exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "budget exceeded") {
		t.Errorf("expected 'budget exceeded' error, got: %v", err)
	}

	// Проверяем, что транзакция НЕ добавлена
	txs = l.ListTransactions()
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction after failed add, got %d", len(txs))
	}
}

// TestSetBudgetValidation — тест валидации бюджета в SetBudget
func TestSetBudgetValidation(t *testing.T) {
	l := NewLedger()
	defer l.Reset()

	// Невалидный бюджет
	budget := Budget{Category: "", Limit: -100.0}
	err := l.SetBudget(budget)
	if err == nil {
		t.Fatal("expected error for invalid budget, got nil")
	}

	// Валидный бюджет
	budget = Budget{Category: "Transport", Limit: 2000.0}
	if err := l.SetBudget(budget); err != nil {
		t.Fatalf("SetBudget failed: %v", err)
	}

	// Проверяем, что бюджет добавлен
	budgets := l.ListBudgets()
	if _, exists := budgets["Transport"]; !exists {
		t.Error("expected 'Transport' budget to exist")
	}
}