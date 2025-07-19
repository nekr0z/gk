package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nekr0z/gk/internal/server/user"
)

const (
	createUserTableQuery = `CREATE TABLE IF NOT EXISTS users (
		username TEXT NOT NULL UNIQUE PRIMARY KEY,
		password BYTEA NOT NULL
	)`
	addUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
	getUserQuery = `SELECT password FROM users WHERE username = $1`
)

var _ user.UserStorage = DB{}

// AddUser adds a user to the database.
func (db DB) AddUser(ctx context.Context, u *user.User) error {
	_, err := db.ExecContext(ctx, addUserQuery, u.Username, u.Password)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return user.ErrAlreadyExists
			}
		}

		return err
	}
	return nil
}

// GetUser retrieves a user from the database.
func (db DB) GetUser(ctx context.Context, username string) (*user.User, error) {
	var password []byte

	err := db.QueryRowContext(ctx, getUserQuery, username).Scan(&password)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user.User{
		Username: username,
		Password: password,
	}, nil
}
