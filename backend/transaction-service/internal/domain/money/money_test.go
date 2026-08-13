package money

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("valid amount", func(t *testing.T) {
		m, err := New(100.5, USD)
		require.NoError(t, err)
		assert.Equal(t, 100.5, m.Amount)
		assert.Equal(t, USD, m.Currency)
	})

	t.Run("zero amount", func(t *testing.T) {
		m, err := New(0, USD)
		require.NoError(t, err)
		assert.Equal(t, 0.0, m.Amount)
		assert.Equal(t, USD, m.Currency)
	})

	t.Run("negative amount", func(t *testing.T) {
		_, err := New(-10, USD)
		assert.ErrorIs(t, err, ErrNegativeAmount)
	})

	t.Run("empty currency", func(t *testing.T) {
		_, err := New(10, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "currency is required")
	})
}

func TestMoney_Add(t *testing.T) {
	m1, _ := New(100, USD)
	m2, _ := New(50, USD)
	m3, _ := New(30, EUR)

	t.Run("same currency", func(t *testing.T) {
		res, err := m1.Add(m2)
		require.NoError(t, err)
		assert.Equal(t, 150.0, res.Amount)
		assert.Equal(t, USD, res.Currency)
	})

	t.Run("different currency", func(t *testing.T) {
		_, err := m1.Add(m3)
		assert.ErrorIs(t, err, ErrCurrencyMismatch)
	})
}

func TestMoney_Sub(t *testing.T) {
	m1, _ := New(100, USD)
	m2, _ := New(30, USD)
	m3, _ := New(30, EUR)

	t.Run("sufficient balance", func(t *testing.T) {
		res, err := m1.Sub(m2)
		require.NoError(t, err)
		assert.Equal(t, 70.0, res.Amount)
		assert.Equal(t, USD, res.Currency)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		_, err := m2.Sub(m1)
		assert.ErrorIs(t, err, ErrInsufficientBalance)
	})

	t.Run("currency mismatch", func(t *testing.T) {
		_, err := m1.Sub(m3)
		assert.ErrorIs(t, err, ErrCurrencyMismatch)
	})
}

func TestMoney_String(t *testing.T) {
	m, _ := New(123.45, USD)
	assert.Equal(t, "123.45 USD", m.String())
}
