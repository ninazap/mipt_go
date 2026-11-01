package main

import (
	"fmt"
	"gateway/internal/api"
	"ledger"
	"net/http"
)

func main() {
	// Инициализация Ledger
	l := ledger.NewLedger()
	
	// Установка начальных бюджетов
	l.SetBudget(ledger.Budget{Category: "Food", Limit: 5000.0, Period: "month"})
	l.SetBudget(ledger.Budget{Category: "Transport", Limit: 2000.0, Period: "month"})
	
	// Настройка роутера
	handler := api.SetupRouter(l)
	
	fmt.Println("Gateway service started on port 8080")
	fmt.Println("API endpoints available at http://localhost:8080/api")
	
	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
