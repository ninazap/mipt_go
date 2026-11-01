package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

// Validate проверяет корректность полей транзакции
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

// Validate проверяет корректность полей бюджета
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

// SetBudget добавляет или обновляет бюджет с валидацией
func (l *Ledger) SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}
	l.budgets[b.Category] = b
	return nil
}

// AddTransaction добавляет транзакцию с валидацией
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

// CheckValid - демонстрация работы с интерфейсом Validatable
func CheckValid(v Validatable) error {
	return v.Validate()
}

func main() {
	ledger := NewLedger()

	fmt.Println("=== Testing Validatable Interface ===\n")

	txValid := Transaction{Amount: 100.0, Category: "Food", Description: "Test"}
	txInvalid := Transaction{Amount: -50.0, Category: "", Description: "Invalid"}

	budgetValid := Budget{Category: "Food", Limit: 5000.0, Period: "month"}
	budgetInvalid := Budget{Category: "", Limit: -100.0, Period: "month"}

	fmt.Println("Testing Transaction validation via interface:")
	if err := CheckValid(txValid); err != nil {
		fmt.Printf("  Valid tx failed: %v\n", err)
	} else {
		fmt.Println("  ✓ Valid transaction passed validation")
	}

	if err := CheckValid(txInvalid); err != nil {
		fmt.Printf("  ✓ Invalid transaction rejected: %v\n", err)
	} else {
		fmt.Println("  ✗ Invalid transaction should have been rejected")
	}

	fmt.Println("\nTesting Budget validation via interface:")
	if err := CheckValid(budgetValid); err != nil {
		fmt.Printf("  Valid budget failed: %v\n", err)
	} else {
		fmt.Println("  ✓ Valid budget passed validation")
	}

	if err := CheckValid(budgetInvalid); err != nil {
		fmt.Printf("  ✓ Invalid budget rejected: %v\n", err)
	} else {
		fmt.Println("  ✗ Invalid budget should have been rejected")
	}

	fmt.Println("\n=== Ledger Operations with Validation ===\n")

	if err := ledger.SetBudget(Budget{Category: "Food", Limit: 5000.0, Period: "month"}); err != nil {
		fmt.Printf("Error setting budget: %v\n", err)
	} else {
		fmt.Println("Budget 'Food' set successfully")
	}

	if err := ledger.SetBudget(Budget{Category: "Transport", Limit: 2000.0, Period: "month"}); err != nil {
		fmt.Printf("Error setting budget: %v\n", err)
	} else {
		fmt.Println("Budget 'Transport' set successfully")
	}

	if err := ledger.SetBudget(Budget{Category: "", Limit: 1000.0}); err != nil {
		fmt.Printf("✓ Expected error for invalid budget: %v\n", err)
	}

	filename := "budgets.json"
	initialBudgetsJSON := `[{"Category": "Entertainment", "Limit": 3000.0, "Period": "month"}]`
	err := os.WriteFile(filename, []byte(initialBudgetsJSON), 0644)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		return
	}
	defer os.Remove(filename)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
	} else {
		defer file.Close()
		reader := bufio.NewReader(file)
		if err := ledger.LoadBudgets(reader); err != nil {
			fmt.Printf("Error loading budgets: %v\n", err)
		} else {
			fmt.Println("Budgets loaded successfully from", filename)
		}
	}

	fmt.Println("\n--- Testing Transactions ---")

	tx1 := Transaction{Amount: 1500.0, Category: "Food", Description: "Grocery shopping"}
	if err := ledger.AddTransaction(tx1); err != nil {
		fmt.Printf("Error adding tx1: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: tx1 added (Amount: %.2f, Category: %s)\n", tx1.Amount, tx1.Category)
	}

	tx2 := Transaction{Amount: 500.0, Category: "Food", Description: "Cafe"}
	if err := ledger.AddTransaction(tx2); err != nil {
		fmt.Printf("Error adding tx2: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: tx2 added (Amount: %.2f, Category: %s)\n", tx2.Amount, tx2.Category)
	}

	tx3 := Transaction{Amount: 3500.0, Category: "Food", Description: "Expensive dinner"}
	if err := ledger.AddTransaction(tx3); err != nil {
		fmt.Printf("EXPECTED ERROR: %v (Category: %s, Amount: %.2f)\n", err, tx3.Category, tx3.Amount)
	} else {
		fmt.Println("UNEXPECTED: tx3 added successfully.")
	}

	txInvalidAmount := Transaction{Amount: 0, Category: "Test", Description: "Invalid amount"}
	if err := ledger.AddTransaction(txInvalidAmount); err != nil {
		fmt.Printf("✓ Expected validation error: %v\n", err)
	}

	txInvalidCategory := Transaction{Amount: 100.0, Category: "", Description: "Empty category"}
	if err := ledger.AddTransaction(txInvalidCategory); err != nil {
		fmt.Printf("✓ Expected validation error: %v\n", err)
	}

	fmt.Println("\n--- All Saved Transactions ---")
	allTransactions := ledger.ListTransactions()
	for _, tx := range allTransactions {
		fmt.Printf("ID: %d, Amount: %.2f, Category: %-12s, Description: %-20s, Date: %s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date)
	}

	fmt.Printf("\nTotal transactions saved: %d\n", len(allTransactions))
}