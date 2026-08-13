package account

import (
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
)

// Account представляет счёт пользователя
type Account struct {
	ID        string
	UserID    string
	Balance   money.Money
	Currency  money.Currency
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewAccount создаёт новый счёт с нулевым балансом
func NewAccount(id, userID string, currency money.Currency) (*Account, error) {
	zeroMoney, err := money.New(0, currency)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:        id,
		UserID:    userID,
		Balance:   zeroMoney,
		Currency:  currency,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// CanWithdraw проверяет, достаточно ли средств для списания
func (a *Account) CanWithdraw(amount money.Money) bool {
	if a.Balance.Currency != amount.Currency {
		return false
	}
	return a.Balance.Amount >= amount.Amount
}

// Withdraw списывает средства (мутирует баланс)
func (a *Account) Withdraw(amount money.Money) error {
	// Проверка валюты
	if a.Balance.Currency != amount.Currency {
		return money.ErrCurrencyMismatch
	}
	// Проверка баланса
	if !a.CanWithdraw(amount) {
		return money.ErrInsufficientBalance
	}
	newBalance, err := a.Balance.Sub(amount)
	if err != nil {
		return err
	}
	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}

// Deposit зачисляет средства
func (a *Account) Deposit(amount money.Money) error {
	if a.Balance.Currency != amount.Currency {
		return money.ErrCurrencyMismatch
	}
	newBalance, err := a.Balance.Add(amount)
	if err != nil {
		return err
	}
	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}
