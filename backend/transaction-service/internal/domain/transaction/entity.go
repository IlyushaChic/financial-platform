package transaction

import (
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
)

// Status определяет статус транзакции
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Transaction представляет финансовую транзакцию
type Transaction struct {
	ID            string
	FromAccountID string // пусто для депозита
	ToAccountID   string
	Amount        money.Money
	Status        Status
	Description   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewTransaction создаёт новую транзакцию со статусом pending
func NewTransaction(id, fromAccountID, toAccountID string, amount money.Money, description string) *Transaction {
	return &Transaction{
		ID:            id,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Status:        StatusPending,
		Description:   description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Complete переводит транзакцию в статус completed
func (t *Transaction) Complete() {
	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

// Fail переводит транзакцию в статус failed
func (t *Transaction) Fail() {
	t.Status = StatusFailed
	t.UpdatedAt = time.Now()
}
