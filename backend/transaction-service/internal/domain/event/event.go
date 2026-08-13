package event

import "time"

// TransactionCompletedEvent публикуется после успешного перевода
type TransactionCompletedEvent struct {
	TransactionID string    `json:"transaction_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Description   string    `json:"description"`
	CompletedAt   time.Time `json:"completed_at"`
}
