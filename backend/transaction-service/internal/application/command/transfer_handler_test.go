package command

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/account"
	eventpkg "github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/event"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/locker"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/transaction"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Mocks ----------

type mockAccountRepo struct {
	getForUpdateFunc func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error)
	updateFunc       func(ctx context.Context, tx *sql.Tx, a *account.Account) error
}

func (m *mockAccountRepo) GetAccountForUpdate(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
	if m.getForUpdateFunc != nil {
		return m.getForUpdateFunc(ctx, tx, id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockAccountRepo) Update(ctx context.Context, tx *sql.Tx, a *account.Account) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, tx, a)
	}
	return nil
}
func (m *mockAccountRepo) Create(ctx context.Context, a *account.Account) error { return nil }

type mockTransactionRepo struct {
	createFunc func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error
	updateFunc func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error
}

func (m *mockTransactionRepo) Create(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, tx, tr)
	}
	return nil
}
func (m *mockTransactionRepo) Update(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, tx, tr)
	}
	return nil
}

// mockLock реализует интерфейс locker.Lock
type mockLock struct {
	unlockFunc func(ctx context.Context) error
}

func (l *mockLock) Unlock(ctx context.Context) error {
	if l.unlockFunc != nil {
		return l.unlockFunc(ctx)
	}
	return nil
}

type mockLocker struct {
	lockFunc func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error)
}

func (m *mockLocker) Lock(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
	if m.lockFunc != nil {
		return m.lockFunc(ctx, key, ttl)
	}
	return &mockLock{}, nil
}

type mockPublisher struct {
	publishFunc func(ctx context.Context, event interface{}) error
}

func (m *mockPublisher) Publish(ctx context.Context, event interface{}) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, event)
	}
	return nil
}

// ---------- Helper ----------
func setupTestHandler(
	accountRepo account.Repository,
	transactionRepo transaction.Repository,
	locker locker.Locker,
	publisher messaging.EventPublisher,
	db *sql.DB,
) *TransferHandler {
	return NewTransferHandler(accountRepo, transactionRepo, locker, db, publisher)
}

// ---------- Tests ----------

func TestTransferHandler_Success(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"
	amount := 50.0
	currency := "USD"
	desc := "test transfer"

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.USD)
	deposit, _ := money.New(100, money.USD)
	fromAcc.Deposit(deposit)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, a *account.Account) error {
			return nil
		},
	}
	transactionRepo := &mockTransactionRepo{
		createFunc: func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
			return nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
			return nil
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	publisher := &mockPublisher{
		publishFunc: func(ctx context.Context, event interface{}) error {
			return nil
		},
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	handler := setupTestHandler(accountRepo, transactionRepo, locker, publisher, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Currency:      currency,
		Description:   desc,
	}

	txEntity, err := handler.Handle(ctx, cmd)
	require.NoError(t, err)
	assert.Equal(t, transaction.StatusCompleted, txEntity.Status)
	assert.Equal(t, amount, txEntity.Amount.Amount)
	assert.Equal(t, money.Currency(currency), txEntity.Amount.Currency)
	assert.Equal(t, desc, txEntity.Description)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_InsufficientBalance(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.USD)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()

	handler := setupTestHandler(accountRepo, &mockTransactionRepo{}, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        50,
		Currency:      "USD",
	}

	_, err = handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.ErrorIs(t, err, money.ErrInsufficientBalance)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_LockError(t *testing.T) {
	ctx := context.Background()
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return nil, errors.New("redis lock error")
		},
	}
	db, _, _ := sqlmock.New()
	handler := setupTestHandler(&mockAccountRepo{}, &mockTransactionRepo{}, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: "from-1",
		ToAccountID:   "to-1",
		Amount:        10,
		Currency:      "USD",
	}
	_, err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to acquire lock")
}

func TestTransferHandler_SameAccount(t *testing.T) {
	ctx := context.Background()
	db, _, _ := sqlmock.New()
	handler := setupTestHandler(&mockAccountRepo{}, &mockTransactionRepo{}, &mockLocker{}, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: "same",
		ToAccountID:   "same",
		Amount:        10,
		Currency:      "USD",
	}
	_, err := handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot transfer to same account")
}

func TestTransferHandler_CurrencyMismatch(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.EUR)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()

	handler := setupTestHandler(accountRepo, &mockTransactionRepo{}, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        10,
		Currency:      "USD",
	}
	_, err = handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "currency mismatch between accounts")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_GetAccountError(t *testing.T) {
	ctx := context.Background()
	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			return nil, errors.New("db connection error")
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()

	handler := setupTestHandler(accountRepo, &mockTransactionRepo{}, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: "from-1",
		ToAccountID:   "to-1",
		Amount:        10,
		Currency:      "USD",
	}
	_, err = handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get from account")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_UpdateAccountError(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.USD)
	deposit, _ := money.New(100, money.USD)
	fromAcc.Deposit(deposit)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, a *account.Account) error {
			return errors.New("update failed")
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()

	handler := setupTestHandler(accountRepo, &mockTransactionRepo{}, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        10,
		Currency:      "USD",
	}
	_, err = handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_CreateTransactionError(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.USD)
	deposit, _ := money.New(100, money.USD)
	fromAcc.Deposit(deposit)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, a *account.Account) error {
			return nil
		},
	}
	transactionRepo := &mockTransactionRepo{
		createFunc: func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
			return errors.New("insert transaction failed")
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()

	handler := setupTestHandler(accountRepo, transactionRepo, locker, &mockPublisher{}, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        10,
		Currency:      "USD",
	}
	_, err = handler.Handle(ctx, cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insert transaction failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferHandler_EventPublished(t *testing.T) {
	ctx := context.Background()
	fromID := "from-1"
	toID := "to-1"
	amount := 50.0

	fromAcc, _ := account.NewAccount(fromID, "user1", money.USD)
	toAcc, _ := account.NewAccount(toID, "user2", money.USD)
	deposit, _ := money.New(100, money.USD)
	fromAcc.Deposit(deposit)

	accountRepo := &mockAccountRepo{
		getForUpdateFunc: func(ctx context.Context, tx *sql.Tx, id string) (*account.Account, error) {
			if id == fromID {
				return fromAcc, nil
			}
			return toAcc, nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, a *account.Account) error {
			return nil
		},
	}
	transactionRepo := &mockTransactionRepo{
		createFunc: func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
			return nil
		},
		updateFunc: func(ctx context.Context, tx *sql.Tx, tr *transaction.Transaction) error {
			return nil
		},
	}
	locker := &mockLocker{
		lockFunc: func(ctx context.Context, key string, ttl time.Duration) (locker.Lock, error) {
			return &mockLock{}, nil
		},
	}
	publisher := &mockPublisher{
		publishFunc: func(ctx context.Context, event interface{}) error {
			ev, ok := event.(eventpkg.TransactionCompletedEvent)
			assert.True(t, ok)
			assert.Equal(t, fromID, ev.FromAccountID)
			assert.Equal(t, toID, ev.ToAccountID)
			assert.Equal(t, amount, ev.Amount)
			return nil
		},
	}
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectCommit()

	handler := setupTestHandler(accountRepo, transactionRepo, locker, publisher, db)
	cmd := TransferCommand{
		FromAccountID: fromID,
		ToAccountID:   toID,
		Amount:        amount,
		Currency:      "USD",
		Description:   "test",
	}
	_, err = handler.Handle(ctx, cmd)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	assert.NoError(t, mock.ExpectationsWereMet())
}
