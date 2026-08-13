package command

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/account"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/event"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/locker"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/transaction"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging"
	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/infrastructure/messaging/rabbitmq"
	"github.com/google/uuid"
)

// TransferCommand ...
type TransferCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        float64
	Currency      string
	Description   string
}

// TransferHandler ...
type TransferHandler struct {
	accountRepo          account.Repository
	transactionRepo      transaction.Repository
	locker               locker.Locker
	db                   *sql.DB
	eventPublisher       messaging.EventPublisher
	notificationProducer *rabbitmq.Producer
}

// NewTransferHandler ...
func NewTransferHandler(
	accountRepo account.Repository,
	transactionRepo transaction.Repository,
	locker locker.Locker,
	db *sql.DB,
	eventPublisher messaging.EventPublisher,
	notificationProducer *rabbitmq.Producer,
) *TransferHandler {
	return &TransferHandler{
		accountRepo:          accountRepo,
		transactionRepo:      transactionRepo,
		locker:               locker,
		db:                   db,
		eventPublisher:       eventPublisher,
		notificationProducer: notificationProducer,
	}
}

// Handle ...
func (h *TransferHandler) Handle(ctx context.Context, cmd TransferCommand) (*transaction.Transaction, error) {
	// 1. Валидация
	if cmd.FromAccountID == cmd.ToAccountID {
		return nil, errors.New("cannot transfer to same account")
	}
	if cmd.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	currency := money.Currency(cmd.Currency)
	amountMoney, err := money.New(cmd.Amount, currency)
	if err != nil {
		return nil, err
	}

	// 2. Блокировка
	lockKey := fmt.Sprintf("account:%s", cmd.FromAccountID)
	lock, err := h.locker.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer lock.Unlock(ctx)

	// 3. Транзакция БД
	tx, err := h.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 4. Получаем счета с FOR UPDATE
	fromAcc, err := h.accountRepo.GetAccountForUpdate(ctx, tx, cmd.FromAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from account: %w", err)
	}
	toAcc, err := h.accountRepo.GetAccountForUpdate(ctx, tx, cmd.ToAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to account: %w", err)
	}

	// 5. Проверка валют и баланса
	if fromAcc.Balance.Currency != toAcc.Balance.Currency {
		return nil, errors.New("currency mismatch between accounts")
	}
	if fromAcc.Balance.Currency != currency {
		return nil, fmt.Errorf("currency mismatch: account has %s, transfer is %s", fromAcc.Balance.Currency, currency)
	}
	if !fromAcc.CanWithdraw(amountMoney) {
		return nil, money.ErrInsufficientBalance
	}

	// 6. Обновление балансов
	if err := fromAcc.Withdraw(amountMoney); err != nil {
		return nil, err
	}
	if err := toAcc.Deposit(amountMoney); err != nil {
		return nil, err
	}

	// 7. Сохранение в БД
	if err := h.accountRepo.Update(ctx, tx, fromAcc); err != nil {
		return nil, err
	}
	if err := h.accountRepo.Update(ctx, tx, toAcc); err != nil {
		return nil, err
	}

	txEntity := transaction.NewTransaction(
		uuid.New().String(),
		cmd.FromAccountID,
		cmd.ToAccountID,
		amountMoney,
		cmd.Description,
	)
	if err := h.transactionRepo.Create(ctx, tx, txEntity); err != nil {
		return nil, err
	}

	// 8. Коммит
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// 9. После коммита: обновляем статус в памяти и в БД (опционально)
	txEntity.Complete()
	updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	updateTx, err := h.db.BeginTx(updateCtx, nil)
	if err == nil {
		_ = h.transactionRepo.Update(updateCtx, updateTx, txEntity)
		_ = updateTx.Commit()
	}

	// 10. Публикация события Kafka (асинхронно)
	go func() {
		evt := event.TransactionCompletedEvent{
			TransactionID: txEntity.ID,
			FromAccountID: txEntity.FromAccountID,
			ToAccountID:   txEntity.ToAccountID,
			Amount:        txEntity.Amount.Amount,
			Currency:      string(txEntity.Amount.Currency),
			Description:   txEntity.Description,
			CompletedAt:   time.Now(),
		}
		_ = h.eventPublisher.Publish(context.Background(), evt)
	}()

	// 11. Отправка уведомления в RabbitMQ (асинхронно)
	go func() {
		routingKey := "notification.email"
		notificationData := map[string]interface{}{
			"user_id":    "user-" + cmd.FromAccountID,
			"amount":     cmd.Amount,
			"currency":   cmd.Currency,
			"type":       "transfer",
			"timestamp":  time.Now().Format(time.RFC3339),
			"to_account": cmd.ToAccountID,
		}
		if err := h.notificationProducer.Publish(context.Background(), routingKey, notificationData); err != nil {
			// логируем ошибку (здесь можно добавить логгер)
		}
	}()

	return txEntity, nil
}
