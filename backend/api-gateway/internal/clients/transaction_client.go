package clients

import (
	"context"

	txpb "github.com/IlyushaChic/financial-platform/backend/transaction-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TransactionClient struct {
	conn   *grpc.ClientConn
	client txpb.TransactionServiceClient
}

func NewTransactionClient(addr string) (*TransactionClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &TransactionClient{
		conn:   conn,
		client: txpb.NewTransactionServiceClient(conn),
	}, nil
}

func (c *TransactionClient) Close() error {
	return c.conn.Close()
}

func (c *TransactionClient) Transfer(ctx context.Context, from, to string, amount float64, currency, description string) (*txpb.TransferResponse, error) {
	req := &txpb.TransferRequest{
		FromAccountId: from,
		ToAccountId:   to,
		Amount:        amount,
		Currency:      currency,
		Description:   description,
	}
	return c.client.Transfer(ctx, req)
}

func (c *TransactionClient) Deposit(ctx context.Context, accountID string, amount float64, currency, description string) (*txpb.DepositResponse, error) {
	req := &txpb.DepositRequest{
		AccountId:   accountID,
		Amount:      amount,
		Currency:    currency,
		Description: description,
	}
	return c.client.Deposit(ctx, req)
}

func (c *TransactionClient) GetBalance(ctx context.Context, accountID string) (*txpb.GetBalanceResponse, error) {
	req := &txpb.GetBalanceRequest{AccountId: accountID}
	return c.client.GetBalance(ctx, req)
}

func (c *TransactionClient) GetTransactions(ctx context.Context, limit, offset int32) (*txpb.GetTransactionsResponse, error) {
	req := &txpb.GetTransactionsRequest{Limit: limit, Offset: offset}
	return c.client.GetTransactions(ctx, req)
}
