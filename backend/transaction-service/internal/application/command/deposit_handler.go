package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/account"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/event"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/transaction"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging"
	"github.com/google/uuid"
)

// DepositCommand содержит данные для пополнения
type DepositCommand struct {
	AccountID   string
	Amount      float64
	Currency    string
	Description string
}

// DepositHandler обрабатывает пополнение счёта
type DepositHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
	db              *sql.DB
	publisher       messaging.EventPublisher
}

// NewDepositHandler создаёт новый экземпляр
func NewDepositHandler(
	accountRepo account.Repository,
	transactionRepo transaction.Repository,
	db *sql.DB,
	publisher messaging.EventPublisher,
) *DepositHandler {
	return &DepositHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		db:              db,
		publisher:       publisher,
	}
}

// Handle выполняет пополнение счёта
func (h *DepositHandler) Handle(ctx context.Context, cmd DepositCommand) (*transaction.Transaction, error) {
	// ----- 1. Валидация -----
	if cmd.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	currency := money.Currency(cmd.Currency)
	amountMoney, err := money.New(cmd.Amount, currency)
	if err != nil {
		return nil, err
	}

	// ----- 2. Транзакция БД -----
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// ----- 3. Получаем счёт с FOR UPDATE (для безопасности) -----
	acc, err := h.accountRepo.GetAccountForUpdate(ctx, tx, cmd.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// ----- 4. Проверка валюты -----
	if acc.Balance.Currency != currency {
		return nil, fmt.Errorf("currency mismatch: account has %s, deposit is %s", acc.Balance.Currency, currency)
	}

	// ----- 5. Зачисление -----
	if err := acc.Deposit(amountMoney); err != nil {
		return nil, err
	}

	// ----- 6. Сохранение -----
	if err := h.accountRepo.Update(ctx, tx, acc); err != nil {
		return nil, err
	}

	// Создаём транзакцию (сразу completed, т.к. операция успешна)
	txEntity := transaction.NewTransaction(
		uuid.New().String(),
		"", // нет отправителя
		cmd.AccountID,
		amountMoney,
		cmd.Description,
	)
	txEntity.Complete() // депозит сразу завершён

	if err := h.transactionRepo.Create(ctx, tx, txEntity); err != nil {
		return nil, err
	}

	// ----- 7. Коммит -----
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// ----- 8. Публикация события -----
	event := event.TransactionCompletedEvent{
		TransactionID: txEntity.ID,
		FromAccountID: "",
		ToAccountID:   cmd.AccountID,
		Amount:        txEntity.Amount.Amount,
		Currency:      string(txEntity.Amount.Currency),
		Description:   txEntity.Description,
		CompletedAt:   time.Now(),
	}
	go func() {
		_ = h.publisher.Publish(context.Background(), event)
	}()

	return txEntity, nil
}
