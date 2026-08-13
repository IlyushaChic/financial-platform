package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/account"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/transaction"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAccountRepository_Integration(t *testing.T) {
	ctx := context.Background()

	// Запускаем PostgreSQL контейнер через GenericContainer
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	defer db.Close()

	// Применяем миграции (создаём таблицы)
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			balance DECIMAL(18,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			from_account_id UUID,
			to_account_id UUID,
			amount DECIMAL(18,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			status VARCHAR(20) NOT NULL,
			description TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	require.NoError(t, err)

	// Создаём репозитории
	accountRepo := NewAccountRepo(db)
	transactionRepo := NewTransactionRepo(db)

	// Подготовка данных
	userID := uuid.New().String()
	acc1, _ := account.NewAccount(uuid.New().String(), userID, money.USD)
	acc2, _ := account.NewAccount(uuid.New().String(), userID, money.USD)

	// Сохраняем счета
	err = accountRepo.Create(ctx, acc1)
	require.NoError(t, err)
	err = accountRepo.Create(ctx, acc2)
	require.NoError(t, err)

	// --- Тест GetAccountForUpdate ---
	t.Run("GetAccountForUpdate", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		fetched, err := accountRepo.GetAccountForUpdate(ctx, tx, acc1.ID)
		require.NoError(t, err)
		assert.Equal(t, acc1.ID, fetched.ID)
		assert.Equal(t, acc1.Balance.Amount, fetched.Balance.Amount)
		assert.Equal(t, acc1.Balance.Currency, fetched.Balance.Currency)
	})

	// --- Тест Update ---
	t.Run("Update", func(t *testing.T) {
		depositAmount, _ := money.New(50, money.USD)
		err = acc1.Deposit(depositAmount)
		require.NoError(t, err)

		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		err = accountRepo.Update(ctx, tx, acc1)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		var balance float64
		err = db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1", acc1.ID).Scan(&balance)
		require.NoError(t, err)
		assert.Equal(t, 50.0, balance)
	})

	// --- Тест транзакций ---
	t.Run("CreateTransaction", func(t *testing.T) {
		amount, _ := money.New(30, money.USD)
		txEntity := transaction.NewTransaction(
			uuid.New().String(),
			acc1.ID,
			acc2.ID,
			amount,
			"test transfer",
		)

		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		defer tx.Rollback()

		err = transactionRepo.Create(ctx, tx, txEntity)
		require.NoError(t, err)

		err = tx.Commit()
		require.NoError(t, err)

		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM transactions WHERE id = $1", txEntity.ID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	// --- Тест конкурентного обновления с FOR UPDATE ---
	t.Run("ConcurrentUpdate", func(t *testing.T) {
		acc3, _ := account.NewAccount(uuid.New().String(), userID, money.USD)
		err = accountRepo.Create(ctx, acc3)
		require.NoError(t, err)

		done := make(chan bool, 2)
		go func() {
			tx, _ := db.BeginTx(ctx, nil)
			defer tx.Rollback()
			fetched, _ := accountRepo.GetAccountForUpdate(ctx, tx, acc3.ID)
			time.Sleep(100 * time.Millisecond)
			fetched.Balance, _ = money.New(100, money.USD)
			accountRepo.Update(ctx, tx, fetched)
			tx.Commit()
			done <- true
		}()
		go func() {
			tx, _ := db.BeginTx(ctx, nil)
			defer tx.Rollback()
			fetched, _ := accountRepo.GetAccountForUpdate(ctx, tx, acc3.ID)
			fetched.Balance, _ = money.New(200, money.USD)
			accountRepo.Update(ctx, tx, fetched)
			tx.Commit()
			done <- true
		}()
		<-done
		<-done

		var balance float64
		err = db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1", acc3.ID).Scan(&balance)
		require.NoError(t, err)
		assert.NotEqual(t, 0.0, balance)
	})
}
