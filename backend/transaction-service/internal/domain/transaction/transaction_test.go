package transaction

import (
	"testing"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/transaction-service/internal/domain/money"
	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	amount, _ := money.New(100, money.USD)
	tx := NewTransaction("tx-1", "from-1", "to-1", amount, "test")

	assert.Equal(t, "tx-1", tx.ID)
	assert.Equal(t, "from-1", tx.FromAccountID)
	assert.Equal(t, "to-1", tx.ToAccountID)
	assert.Equal(t, amount, tx.Amount)
	assert.Equal(t, StatusPending, tx.Status)
	assert.Equal(t, "test", tx.Description)
	assert.NotZero(t, tx.CreatedAt)
	assert.NotZero(t, tx.UpdatedAt)
}

func TestTransaction_Complete(t *testing.T) {
	amount, _ := money.New(50, money.USD)
	tx := NewTransaction("tx-1", "from-1", "to-1", amount, "")

	tx.Complete()
	assert.Equal(t, StatusCompleted, tx.Status)
	assert.WithinDuration(t, time.Now(), tx.UpdatedAt, time.Second)
}

func TestTransaction_Fail(t *testing.T) {
	amount, _ := money.New(50, money.USD)
	tx := NewTransaction("tx-1", "from-1", "to-1", amount, "")

	tx.Fail()
	assert.Equal(t, StatusFailed, tx.Status)
	assert.WithinDuration(t, time.Now(), tx.UpdatedAt, time.Second)
}
