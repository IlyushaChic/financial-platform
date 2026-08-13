package postgres

import (
	"context"
	"database/sql"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/transaction"
)

type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
	query := `INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, status, description, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	_, err := tx.ExecContext(ctx, query,
		tr.ID,
		tr.FromAccountID,
		tr.ToAccountID,
		tr.Amount.Amount,
		tr.Amount.Currency,
		tr.Status,
		tr.Description,
	)
	return err
}

func (r *TransactionRepo) Update(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
	query := `UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, tr.Status, tr.ID)
	return err
}
