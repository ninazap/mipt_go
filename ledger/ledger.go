package ledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"ledger/cache"
	"ledger/db"
)

// Validatable interface - контракт для проверки валидности данных
type Validatable interface {
	Validate() error
}

type Transaction struct {
	ID          int     `json:"id"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
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
	Category string  `json:"category"`
	Limit    float64 `json:"limit"`
	Period   string  `json:"period"`
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
	db    *db.Database
	cache *cache.Cache
}

func NewLedger(database *db.Database, cache *cache.Cache) *Ledger {
	return &Ledger{
		db:    database,
		cache: cache,
	}
}

func (l *Ledger) SetBudget(b Budget) error {
	if err := b.Validate(); err != nil {
		return fmt.Errorf("invalid budget: %w", err)
	}

	// UPSERT бюджета
	query := `
		INSERT INTO budgets (category, limit_amount) 
		VALUES ($1, $2) 
		ON CONFLICT (category) 
		DO UPDATE SET limit_amount = EXCLUDED.limit_amount, updated_at = CURRENT_TIMESTAMP
	`
	_, err := l.db.Exec(query, b.Category, b.Limit)
	if err != nil {
		return fmt.Errorf("failed to set budget: %w", err)
	}

	// Инвалидация кеша бюджетов
	if l.cache != nil {
		l.cache.Delete("budgets:all")
	}

	return nil
}

func (l *Ledger) AddTransaction(tx Transaction) error {
	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Проверка бюджета
	var limit float64
	err := l.db.QueryRow("SELECT limit_amount FROM budgets WHERE category = $1", tx.Category).Scan(&limit)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check budget: %w", err)
	}

	// Если бюджет существует, проверяем лимит
	if err == nil {
		var spent float64
		err = l.db.QueryRow("SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = $1", tx.Category).Scan(&spent)
		if err != nil {
			return fmt.Errorf("failed to calculate spent amount: %w", err)
		}

		if spent+tx.Amount > limit {
			return errors.New("budget exceeded")
		}
	}

	// Вставка транзакции
	var id int
	err = l.db.QueryRow(
		"INSERT INTO expenses (amount, category, description, date) VALUES ($1, $2, $3, $4) RETURNING id",
		tx.Amount, tx.Category, tx.Description, tx.Date,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to add transaction: %w", err)
	}

	tx.ID = id
	return nil
}

func (l *Ledger) ListTransactions() ([]Transaction, error) {
	rows, err := l.db.Query(`
		SELECT id, amount, category, description, date 
		FROM expenses 
		ORDER BY date DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		if err := rows.Scan(&tx.ID, &tx.Amount, &tx.Category, &tx.Description, &tx.Date); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if transactions == nil {
		transactions = []Transaction{}
	}

	return transactions, nil
}

func (l *Ledger) ListBudgets() ([]Budget, error) {
	// Попытка получить из кеша
	if l.cache != nil {
		var budgets []Budget
		if err := l.cache.Get("budgets:all", &budgets); err == nil {
			return budgets, nil
		}
	}

	rows, err := l.db.Query("SELECT category, limit_amount FROM budgets ORDER BY category")
	if err != nil {
		return nil, fmt.Errorf("failed to list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []Budget
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.Category, &b.Limit); err != nil {
			return nil, fmt.Errorf("failed to scan budget: %w", err)
		}
		b.Period = "month" // По умолчанию
		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if budgets == nil {
		budgets = []Budget{}
	}

	// Сохранение в кеш
	if l.cache != nil {
		l.cache.Set("budgets:all", budgets, 30*time.Second)
	}

	return budgets, nil
}

// ReportSummary структура для отчёта
type ReportSummary struct {
	TotalAmount float64           `json:"total_amount"`
	ByCategory  map[string]float64 `json:"by_category"`
	From        string            `json:"from"`
	To          string            `json:"to"`
}

// GetReportSummary получает отчёт с кэшированием
func (l *Ledger) GetReportSummary(from, to string) (*ReportSummary, error) {
	cacheKey := fmt.Sprintf("report:summary:%s:%s", from, to)

	// Попытка получить из кеша
	if l.cache != nil {
		var summary ReportSummary
		if err := l.cache.Get(cacheKey, &summary); err == nil {
			return &summary, nil
		}
	}

	// Расчёт отчёта из БД
	rows, err := l.db.Query(`
		SELECT category, SUM(amount) as total
		FROM expenses
		WHERE date BETWEEN $1 AND $2
		GROUP BY category
	`, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate report: %w", err)
	}
	defer rows.Close()

	summary := &ReportSummary{
		TotalAmount: 0,
		ByCategory:  make(map[string]float64),
		From:        from,
		To:          to,
	}

	for rows.Next() {
		var category string
		var amount float64
		if err := rows.Scan(&category, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		summary.ByCategory[category] = amount
		summary.TotalAmount += amount
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Сохранение в кеш
	if l.cache != nil {
		l.cache.Set(cacheKey, summary, 30*time.Second)
	}

	return summary, nil
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

// Reset очищает состояние для тестов (удаляет данные из БД)
func (l *Ledger) Reset() error {
	_, err := l.db.Exec("DELETE FROM expenses")
	if err != nil {
		return err
	}
	_, err = l.db.Exec("DELETE FROM budgets")
	if err != nil {
		return err
	}
	return nil
}
