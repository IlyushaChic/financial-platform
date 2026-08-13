package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/IlyushaChic/financial-platform/backend/auth-service/internal/domain/user"
	"github.com/lib/pq"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) error {
	query := `INSERT INTO users (id, email, password_hash, full_name, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err := r.db.ExecContext(ctx, query, u.ID, u.Email, u.PasswordHash, u.FullName)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.New("user already exists")
		}
		return err
	}
	return nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, email, password_hash, full_name, created_at, updated_at
			  FROM users WHERE email = $1`
	row := r.db.QueryRowContext(ctx, query, email)
	u := &user.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*user.User, error) {
	query := `SELECT id, email, password_hash, full_name, created_at, updated_at
			  FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	u := &user.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return u, err
}
