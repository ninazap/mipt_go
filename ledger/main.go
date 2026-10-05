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

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

type Budget struct {
	Category string
	Limit    float64
	Period   string
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

func (l *Ledger) SetBudget(b Budget) {
	l.budgets[b.Category] = b
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return errors.New("transaction amount must be greater than zero")
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
		l.SetBudget(b)
	}
	return nil
}

func main() {
	ledger := NewLedger()

	ledger.SetBudget(Budget{Category: "Food", Limit: 5000.0, Period: "month"})
	ledger.SetBudget(Budget{Category: "Transport", Limit: 2000.0, Period: "month"})
	fmt.Println("Initial budgets set via code.")

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

	txInvalid := Transaction{Amount: 0, Category: "Test", Description: "Invalid transaction"}
	if err := ledger.AddTransaction(txInvalid); err != nil {
		fmt.Printf("EXPECTED VALIDATION ERROR: %v\n", err)
	}

	fmt.Println("\n--- All Saved Transactions ---")
	allTransactions := ledger.ListTransactions()
	for _, tx := range allTransactions {
		fmt.Printf("ID: %d, Amount: %.2f, Category: %-12s, Description: %-20s, Date: %s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date)
	}
}
