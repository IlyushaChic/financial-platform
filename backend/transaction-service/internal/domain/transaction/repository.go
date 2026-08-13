package transaction

import (
	"context"
	"database/sql"
)

// Repository определяет методы для работы с транзакциями
type Repository interface {
	// Create создаёт запись транзакции внутри транзакции БД
	Create(ctx context.Context, tx *sql.Tx, tr *Transaction) error
	// Update обновляет статус транзакции
	Update(ctx context.Context, tx *sql.Tx, tr *Transaction) error
}
