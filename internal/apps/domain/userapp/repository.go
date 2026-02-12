package userapp

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
}

type SQLiteUserRepo struct {
	db *sqlx.DB
}

func NewSQLiteUserRepo(db *sqlx.DB) *SQLiteUserRepo {
	return &SQLiteUserRepo{db: db}
}

func (r *SQLiteUserRepo) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User

	query := `SELECT id, tenant_id, email, password_hash, role, created_at
        		FROM users 
        		WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return User{}, err
	}

	return user, nil

}
