package ledger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// Validatable interface - контракт для проверки валидности данных
type Validatable interface {
	Validate() error
}

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

func (tx Transaction) Validate() error {
	if tx.Amount <= 0 {
		return errors.New("transaction amount must be greater than zero")
	}
	if tx.Category == "" {
		return errors.New("transaction category cannot be empty")
	}
	return nil
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
}

func (b Budget) Validate() error {
	if b.Limit <= 0 {
		return errors.New("budget limit must be positive")
	}
	if b.Category == "" {
		return errors.New("budget category cannot be empty")
	}
	return nil
}

type Ledger struct {
	transactions []Transaction
	budgets      map[string]Budget
}

func NewLedger() *Ledger {
	return &Ledger{
		transactions: make([]Transaction, 0),
		budgets:      make(map[string]Budget),
	}
}

func (l *Ledger) SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	l.budgets[b.Category] = b
	return nil
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	if budget, exists := l.budgets[tx.Category]; exists {
		var currentSum float64
		for _, t := range l.transactions {
			if t.Category == tx.Category {
				currentSum += t.Amount
			}
		}
		if currentSum+tx.Amount > budget.Limit {
			return errors.New("budget exceeded")
		}
	}

	tx.ID = len(l.transactions) + 1
	tx.Date = time.Now().Format("2006-01-02")
	l.transactions = append(l.transactions, tx)
	return nil
}

func (l *Ledger) ListTransactions() []Transaction {
	return l.transactions
}

func (l *Ledger) ListBudgets() map[string]Budget {
	return l.budgets
}

func (l *Ledger) LoadBudgets(r io.Reader) error {
	var budgets []Budget
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&budgets); err != nil {
		return fmt.Errorf("failed to parse budgets JSON: %w", err)
	}
	for _, b := range budgets {
		if err := l.SetBudget(b); err != nil {
			return fmt.Errorf("failed to set budget: %w", err)
		}
	}
	return nil
}

func (l *Ledger) Reset() {
	l.transactions = make([]Transaction, 0)
	l.budgets = make(map[string]Budget)
}