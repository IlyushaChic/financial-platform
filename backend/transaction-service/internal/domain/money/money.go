package money

import (
	"errors"
	"fmt"
)

// Currency представляет валюту (ISO 4217)
type Currency string

const (
	USD Currency = "USD"
	EUR Currency = "EUR"
	RUB Currency = "RUB"
)

// Money представляет сумму с валютой
type Money struct {
	Amount   float64
	Currency Currency
}

// Ошибки, которые могут возвращаться при работе с деньгами
var (
	ErrNegativeAmount      = errors.New("amount cannot be negative")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// New создаёт Money с проверкой, что сумма не отрицательная
func New(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		return Money{}, errors.New("currency is required")
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// Add складывает две суммы одной валюты
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Sub вычитает другую сумму (проверяет, что результат не отрицательный)
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	if m.Amount < other.Amount {
		return Money{}, ErrInsufficientBalance
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// String возвращает строковое представление
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}
