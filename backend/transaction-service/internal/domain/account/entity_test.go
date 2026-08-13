package account

import (
	"testing"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccount_Withdraw(t *testing.T) {
	acc, err := NewAccount("1", "user1", money.USD)
	require.NoError(t, err)

	deposit, _ := money.New(100, money.USD)
	err = acc.Deposit(deposit)
	require.NoError(t, err)

	t.Run("successful withdraw", func(t *testing.T) {
		withdrawAmount, _ := money.New(30, money.USD)
		err := acc.Withdraw(withdrawAmount)
		assert.NoError(t, err)
		assert.Equal(t, 70.0, acc.Balance.Amount)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		withdrawAmount, _ := money.New(100, money.USD)
		err := acc.Withdraw(withdrawAmount)
		assert.ErrorIs(t, err, money.ErrInsufficientBalance)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		withdrawAmount, _ := money.New(10, money.EUR)
		err := acc.Withdraw(withdrawAmount)
		assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
	})
}

func TestAccount_Deposit(t *testing.T) {
	acc, err := NewAccount("1", "user1", money.USD)
	require.NoError(t, err)

	t.Run("successful deposit", func(t *testing.T) {
		depositAmount, _ := money.New(50, money.USD)
		err := acc.Deposit(depositAmount)
		assert.NoError(t, err)
		assert.Equal(t, 50.0, acc.Balance.Amount)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		depositAmount, _ := money.New(10, money.EUR)
		err := acc.Deposit(depositAmount)
		assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
	})
}

func TestMoney_Sub(t *testing.T) {
	m1, _ := money.New(100, money.USD)
	m2, _ := money.New(30, money.USD)

	t.Run("sufficient balance", func(t *testing.T) {
		res, err := m1.Sub(m2)
		assert.NoError(t, err)
		assert.Equal(t, 70.0, res.Amount)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		_, err := m2.Sub(m1)
		assert.ErrorIs(t, err, money.ErrInsufficientBalance)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		m3, _ := money.New(10, money.EUR)
		_, err := m1.Sub(m3)
		assert.ErrorIs(t, err, money.ErrCurrencyMismatch)
	})
}
