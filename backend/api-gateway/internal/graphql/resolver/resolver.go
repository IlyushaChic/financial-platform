package resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/clients"
	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/graphql/generated"
)

type Resolver struct {
	authClient        *clients.AuthClient
	transactionClient *clients.TransactionClient
}

func NewResolver(authClient *clients.AuthClient, transactionClient *clients.TransactionClient) *Resolver {
	return &Resolver{
		authClient:        authClient,
		transactionClient: transactionClient,
	}
}

// ---------- Query ----------
func (r *Resolver) Me(ctx context.Context) (*generated.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}
	return &generated.User{
		ID:       userID,
		Email:    "user@example.com",
		FullName: "Test User",
	}, nil
}

func (r *Resolver) Balance(ctx context.Context, accountID string) (*generated.Account, error) {
	// Заглушка (позже заменим на реальный вызов)
	return &generated.Account{
		ID:       accountID,
		UserID:   "user-1",
		Balance:  100.0,
		Currency: "USD",
	}, nil
}

func (r *Resolver) Transactions(ctx context.Context, limit *int, offset *int) ([]*generated.Transaction, error) {
	// Заглушка
	return []*generated.Transaction{
		{
			ID:            "tx-1",
			FromAccountID: "from-1",
			ToAccountID:   "to-1",
			Amount:        50,
			Currency:      "USD",
			Status:        "completed",
			Description:   "test",
			CreatedAt:     time.Now().Format(time.RFC3339),
		},
	}, nil
}

// ---------- Mutation ----------
func (r *Resolver) Register(ctx context.Context, email string, password string, fullName string) (*generated.AuthPayload, error) {
	// Заглушка
	return &generated.AuthPayload{
		AccessToken:  "fake-access-token",
		RefreshToken: "fake-refresh-token",
		ExpiresIn:    3600,
		User: &generated.User{
			ID:       "user-1",
			Email:    email,
			FullName: fullName,
		},
	}, nil
}

func (r *Resolver) Login(ctx context.Context, email string, password string) (*generated.AuthPayload, error) {
	return &generated.AuthPayload{
		AccessToken:  "fake-access-token",
		RefreshToken: "fake-refresh-token",
		ExpiresIn:    3600,
		User: &generated.User{
			ID:       "user-1",
			Email:    email,
			FullName: "User",
		},
	}, nil
}

func (r *Resolver) Transfer(ctx context.Context, toAccountID string, amount float64, currency string, description *string) (*generated.Transaction, error) {
	return &generated.Transaction{
		ID:            "tx-new",
		FromAccountID: "from-1",
		ToAccountID:   toAccountID,
		Amount:        amount,
		Currency:      currency,
		Status:        "pending",
		Description:   "transfer",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

func (r *Resolver) Deposit(ctx context.Context, accountID string, amount float64, currency string, description *string) (*generated.Transaction, error) {
	return &generated.Transaction{
		ID:            "tx-deposit",
		FromAccountID: "",
		ToAccountID:   accountID,
		Amount:        amount,
		Currency:      currency,
		Status:        "completed",
		Description:   "deposit",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

// ---------- Subscription ----------
func (r *Resolver) TransactionCompleted(ctx context.Context) (<-chan *generated.Transaction, error) {
	ch := make(chan *generated.Transaction)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- &generated.Transaction{
			ID:            "test-sub-1",
			FromAccountID: "from-1",
			ToAccountID:   "to-1",
			Amount:        100,
			Currency:      "USD",
			Status:        "completed",
			Description:   "Test subscription",
			CreatedAt:     time.Now().Format(time.RFC3339),
		}
	}()
	return ch, nil
}
