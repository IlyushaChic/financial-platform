package main

import (
	"log"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}
type Account struct {
	ID       string  `json:"id"`
	UserID   string  `json:"userID"`
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
}
type Transaction struct {
	ID            string  `json:"id"`
	FromAccountID string  `json:"fromAccountID"`
	ToAccountID   string  `json:"toAccountID"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Description   string  `json:"description"`
	CreatedAt     string  `json:"createdAt"`
}
type AuthPayload struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	User         User   `json:"user"`
	Message      string `json:"message"`
}

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name: "User",
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.String},
		"email":    &graphql.Field{Type: graphql.String},
		"fullName": &graphql.Field{Type: graphql.String},
	},
})
var accountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"id":       &graphql.Field{Type: graphql.String},
		"userID":   &graphql.Field{Type: graphql.String},
		"balance":  &graphql.Field{Type: graphql.Float},
		"currency": &graphql.Field{Type: graphql.String},
	},
})
var transactionType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Transaction",
	Fields: graphql.Fields{
		"id":            &graphql.Field{Type: graphql.String},
		"fromAccountID": &graphql.Field{Type: graphql.String},
		"toAccountID":   &graphql.Field{Type: graphql.String},
		"amount":        &graphql.Field{Type: graphql.Float},
		"currency":      &graphql.Field{Type: graphql.String},
		"status":        &graphql.Field{Type: graphql.String},
		"description":   &graphql.Field{Type: graphql.String},
		"createdAt":     &graphql.Field{Type: graphql.String},
	},
})
var authPayloadType = graphql.NewObject(graphql.ObjectConfig{
	Name: "AuthPayload",
	Fields: graphql.Fields{
		"accessToken":  &graphql.Field{Type: graphql.String},
		"refreshToken": &graphql.Field{Type: graphql.String},
		"expiresIn":    &graphql.Field{Type: graphql.Int},
		"user":         &graphql.Field{Type: userType},
		"message":      &graphql.Field{Type: graphql.String},
	},
})

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"me": &graphql.Field{
			Type: userType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return User{ID: "user-1", Email: "test@example.com", FullName: "Test User"}, nil
			},
		},
		"balance": &graphql.Field{
			Type: accountType,
			Args: graphql.FieldConfigArgument{
				"accountID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				accountID, _ := p.Args["accountID"].(string)
				return Account{ID: accountID, UserID: "user-1", Balance: 100.5, Currency: "USD"}, nil
			},
		},
		"transactions": &graphql.Field{
			Type: graphql.NewList(transactionType),
			Args: graphql.FieldConfigArgument{
				"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
				"offset": &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return []Transaction{{
					ID: "tx-1", FromAccountID: "from-1", ToAccountID: "to-1",
					Amount: 50, Currency: "USD", Status: "completed",
					Description: "test", CreatedAt: time.Now().Format(time.RFC3339),
				}}, nil
			},
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		"register": &graphql.Field{
			Type: authPayloadType,
			Args: graphql.FieldConfigArgument{
				"email":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"password": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"fullName": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				email, _ := p.Args["email"].(string)
				fullName, _ := p.Args["fullName"].(string)
				return AuthPayload{
					AccessToken: "fake-token", RefreshToken: "fake-refresh", ExpiresIn: 3600,
					User:    User{ID: "user-1", Email: email, FullName: fullName},
					Message: "Registered (mock)",
				}, nil
			},
		},
		"login": &graphql.Field{
			Type: authPayloadType,
			Args: graphql.FieldConfigArgument{
				"email":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"password": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				email, _ := p.Args["email"].(string)
				return AuthPayload{
					AccessToken: "fake-token", RefreshToken: "fake-refresh", ExpiresIn: 3600,
					User:    User{ID: "user-1", Email: email, FullName: "Test"},
					Message: "Logged in (mock)",
				}, nil
			},
		},
		"transfer": &graphql.Field{
			Type: transactionType,
			Args: graphql.FieldConfigArgument{
				"toAccountID": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"amount":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
				"currency":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"description": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				toAccountID, _ := p.Args["toAccountID"].(string)
				amount, _ := p.Args["amount"].(float64)
				currency, _ := p.Args["currency"].(string)
				return Transaction{
					ID: "tx-new", FromAccountID: "from-1", ToAccountID: toAccountID,
					Amount: amount, Currency: currency, Status: "pending",
					Description: "transfer", CreatedAt: time.Now().Format(time.RFC3339),
				}, nil
			},
		},
		"deposit": &graphql.Field{
			Type: transactionType,
			Args: graphql.FieldConfigArgument{
				"accountID":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"amount":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Float)},
				"currency":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				"description": &graphql.ArgumentConfig{Type: graphql.String},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				accountID, _ := p.Args["accountID"].(string)
				amount, _ := p.Args["amount"].(float64)
				currency, _ := p.Args["currency"].(string)
				return Transaction{
					ID: "dep-1", FromAccountID: "", ToAccountID: accountID,
					Amount: amount, Currency: currency, Status: "completed",
					Description: "deposit", CreatedAt: time.Now().Format(time.RFC3339),
				}, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    queryType,
	Mutation: mutationType,
})

func main() {
	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
	cors := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	http.Handle("/", cors(h))
	http.Handle("/metrics", promhttp.Handler())

	log.Println("🚀 Server on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
