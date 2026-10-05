package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	// 1. Вывод имени пользователя из переменной окружения
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") 
	}
	fmt.Printf("Имя пользователя: %s\n", username)

	// 2. Считывание и вывод аргументов CLI
	fmt.Printf("Аргументы CLI: %v\n", os.Args[1:])

	// 3. Вывод версии Go
	fmt.Printf("Версия Go: %s\n", runtime.Version())
}