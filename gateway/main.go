package main

import (
	"fmt"
	"gateway/internal/api"
	"ledger"
	"ledger/cache"
	"ledger/db"
	"net/http"
)

func main() {
	// Инициализация базы данных
	database, err := db.NewDatabase()
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer database.Close()

	// Инициализация Redis кеша
	cacheClient, err := cache.NewCache()
	if err != nil {
		fmt.Printf("Warning: Failed to connect to Redis: %v\n", err)
		fmt.Println("Running without cache...")
	} else {
		defer cacheClient.Close()
	}

	// Инициализация Ledger (переименовали переменную, чтобы не затенять пакет)
	ledgerService := ledger.NewLedger(database, cacheClient)

	// Настройка роутера
	handler := api.SetupRouter(ledgerService)

	fmt.Println("Gateway service started on port 8080")
	fmt.Println("API endpoints available at http://localhost:8080/api")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
