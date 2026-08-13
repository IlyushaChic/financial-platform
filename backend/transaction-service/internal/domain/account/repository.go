package account

import (
	"context"
	"database/sql"
)

// Repository определяет методы для работы со счетами
type Repository interface {
	// GetAccountForUpdate получает счёт с блокировкой FOR UPDATE (внутри транзакции)
	GetAccountForUpdate(ctx context.Context, tx *sql.Tx, accountID string) (*Account, error)
	// Update обновляет счёт внутри транзакции
	Update(ctx context.Context, tx *sql.Tx, account *Account) error
	// Create создаёт новый счёт
	Create(ctx context.Context, account *Account) error
}
