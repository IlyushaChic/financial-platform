package resolver

import (
	"context"
	"time"

	"github.com/IlyushaChic/financial-platform/backend/api-gateway/internal/graphql/generated"
)

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

func (r *mutationResolver) Register(ctx context.Context, email, password, fullName string) (*generated.AuthPayload, error) {
	return &generated.AuthPayload{
		AccessToken:  "fake-token",
		RefreshToken: "fake-refresh",
		ExpiresIn:    3600,
		User:         generated.User{ID: "user-1", Email: email, FullName: fullName},
		Message:      "Registered (mock)",
	}, nil
}

func (r *mutationResolver) Login(ctx context.Context, email, password string) (*generated.AuthPayload, error) {
	return &generated.AuthPayload{
		AccessToken:  "fake-token",
		RefreshToken: "fake-refresh",
		ExpiresIn:    3600,
		User:         generated.User{ID: "user-1", Email: email, FullName: "Test"},
		Message:      "Logged in (mock)",
	}, nil
}

func (r *mutationResolver) Transfer(ctx context.Context, toAccountID string, amount float64, currency string, description *string) (*generated.Transaction, error) {
	return &generated.Transaction{
		ID:            "tx-" + time.Now().Format("20060102150405"),
		FromAccountID: "from-1",
		ToAccountID:   toAccountID,
		Amount:        amount,
		Currency:      currency,
		Status:        "completed",
		Description:   "mock transfer",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

func (r *mutationResolver) Deposit(ctx context.Context, accountID string, amount float64, currency string, description *string) (*generated.Transaction, error) {
	return &generated.Transaction{
		ID:            "dep-" + time.Now().Format("20060102150405"),
		FromAccountID: "",
		ToAccountID:   accountID,
		Amount:        amount,
		Currency:      currency,
		Status:        "completed",
		Description:   "mock deposit",
		CreatedAt:     time.Now().Format(time.RFC3339),
	}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*generated.User, error) {
	return &generated.User{ID: "user-1", Email: "test@example.com", FullName: "Test User"}, nil
}

func (r *queryResolver) Balance(ctx context.Context, accountID string) (*generated.Account, error) {
	return &generated.Account{ID: accountID, UserID: "user-1", Balance: 100.5, Currency: "USD"}, nil
}

func (r *queryResolver) Transactions(ctx context.Context, limit *int, offset *int) ([]*generated.Transaction, error) {
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

func (r *subscriptionResolver) TransactionCompleted(ctx context.Context) (<-chan *generated.Transaction, error) {
	ch := make(chan *generated.Transaction)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- &generated.Transaction{
			ID:            "sub-1",
			FromAccountID: "from-sub",
			ToAccountID:   "to-sub",
			Amount:        200,
			Currency:      "USD",
			Status:        "completed",
			Description:   "subscription mock",
			CreatedAt:     time.Now().Format(time.RFC3339),
		}
	}()
	return ch, nil
}
