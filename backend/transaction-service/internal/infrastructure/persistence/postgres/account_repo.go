package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/account"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
)

type AccountRepo struct {
	db *sql.DB
}

func NewAccountRepo(db *sql.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) GetAccountForUpdate(ctx context.Context, tx *sql.Tx, accountID string) (*account.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at, updated_at
			  FROM accounts WHERE id = $1 FOR UPDATE`
	row := tx.QueryRowContext(ctx, query, accountID)

	var acc account.Account
	var balance float64
	var currency string
	err := row.Scan(&acc.ID, &acc.UserID, &balance, &currency, &acc.CreatedAt, &acc.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("account %s not found", accountID)
	}
	if err != nil {
		return nil, err
	}
	m, err := money.New(balance, money.Currency(currency))
	if err != nil {
		return nil, err
	}
	acc.Balance = m
	acc.Currency = money.Currency(currency)
	return &acc, nil
}

func (r *AccountRepo) Update(ctx context.Context, tx *sql.Tx, acc *account.Account) error {
	query := `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, acc.Balance.Amount, acc.ID)
	return err
}

func (r *AccountRepo) Create(ctx context.Context, acc *account.Account) error {
	query := `INSERT INTO accounts (id, user_id, balance, currency, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, acc.ID, acc.UserID, acc.Balance.Amount, acc.Balance.Currency)
	return err
}
