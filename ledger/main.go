package main

import (
	"fmt"
	"time"
)

type Transaction struct {
	ID          int
	Amount      float64
	Category    string
	Description string
	Date        string
}

var transactions []Transaction

func AddTransaction(tx Transaction) error {
	if tx.Amount == 0 {
		return fmt.Errorf("transaction amount cannot be zero")
	}
	
	tx.ID = len(transactions) + 1
	tx.Date = time.Now().Format("2006-01-02")
	
	transactions = append(transactions, tx)
	return nil
}

func ListTransactions() []Transaction {
	return transactions
}

func main() {
	fmt.Println("Ledger service started")
	
	tx1 := Transaction{
		Amount:      150.50,
		Category:    "Food",
		Description: "Grocery shopping",
	}
	
	tx2 := Transaction{
		Amount:      50.00,
		Category:    "Transport",
		Description: "Taxi ride",
	}
	
	tx3 := Transaction{
		Amount:      200.00,
		Category:    "Entertainment",
		Description: "Concert tickets",
	}
	
	if err := AddTransaction(tx1); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	
	if err := AddTransaction(tx2); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	
	if err := AddTransaction(tx3); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	
	txInvalid := Transaction{
		Amount:      0,
		Category:    "Test",
		Description: "Invalid transaction",
	}
	
	if err := AddTransaction(txInvalid); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}
	
	allTransactions := ListTransactions()
	
	fmt.Println("\nAll transactions:")
	for _, tx := range allTransactions {
		fmt.Printf("ID: %d, Amount: %.2f, Category: %s, Description: %s, Date: %s\n",
			tx.ID, tx.Amount, tx.Category, tx.Description, tx.Date)
	}
}
