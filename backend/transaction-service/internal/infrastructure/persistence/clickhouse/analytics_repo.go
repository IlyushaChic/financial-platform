package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type AnalyticsRepo struct {
	db *sql.DB
}

func NewAnalyticsRepo(dsn string) (*AnalyticsRepo, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("clickhouse ping failed: %w", err)
	}
	return &AnalyticsRepo{db: db}, nil
}

func (r *AnalyticsRepo) Close() error {
	return r.db.Close()
}

// InsertTransaction вставляет запись о транзакции в аналитическую таблицу
func (r *AnalyticsRepo) InsertTransaction(ctx context.Context, txID, fromAccount, toAccount string, amount float64, currency string, status string, createdAt time.Time) error {
	query := `
		INSERT INTO transactions_analytics (
			transaction_id, from_account_id, to_account_id, amount, currency, status, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, txID, fromAccount, toAccount, amount, currency, status, createdAt)
	if err != nil {
		return fmt.Errorf("failed to insert into clickhouse: %w", err)
	}
	return nil
}

// InsertAggregatedData (опционально) можно добавить для агрегации по дням/пользователям
// Для простоты оставим одну таблицу.

// InsertDailyUserAggregate вставляет агрегированные данные по дням и пользователям
func (r *AnalyticsRepo) InsertDailyUserAggregate(ctx context.Context, userID string, date time.Time, totalAmount float64, transactionCount int64, currency string) error {
	query := `
		INSERT INTO daily_user_transactions (user_id, date, total_amount, transaction_count, currency)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, userID, date, totalAmount, transactionCount, currency)
	if err != nil {
		return fmt.Errorf("failed to insert daily user aggregate: %w", err)
	}
	return nil
}

// InsertDailyCurrencyAggregate вставляет агрегированные данные по дням и валютам
func (r *AnalyticsRepo) InsertDailyCurrencyAggregate(ctx context.Context, date time.Time, currency string, totalAmount float64, transactionCount int64) error {
	query := `
		INSERT INTO daily_currency_stats (date, currency, total_amount, transaction_count)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, date, currency, totalAmount, transactionCount)
	if err != nil {
		return fmt.Errorf("failed to insert daily currency aggregate: %w", err)
	}
	return nil
}
