package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nekr0z/gk/internal/hash"
	"github.com/nekr0z/gk/internal/server/secret"
)

const (
	createSecretTableQuery = `CREATE TABLE IF NOT EXISTS secrets (
		username TEXT,
		key TEXT,
		data BYTEA,
		hash BYTEA,
		PRIMARY KEY (username, key)
	)`
	addSecretQuery    = `INSERT INTO secrets (username, key, data, hash) VALUES ($1, $2, $3, $4)`
	updateSecretQuery = `UPDATE secrets SET data = $1, hash = $2 WHERE username = $3 AND key = $4 AND hash = $5`
	getSecretQuery    = `SELECT data, hash FROM secrets WHERE username = $1 AND key = $2`
	deleteSecretQuery = `DELETE FROM secrets WHERE username = $1 AND key = $2 AND hash = $3`
	listHashesQuery   = `SELECT key, hash FROM secrets WHERE username = $1`
)

var _ secret.SecretStorage = DB{}

// Get returns a secret from the database.
func (db DB) Get(ctx context.Context, username, key string) (secret.Secret, error) {
	var (
		data       []byte
		storedHash []byte
	)

	err := db.QueryRowContext(ctx, getSecretQuery, username, key).Scan(&data, &storedHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return secret.Secret{}, secret.ErrNotFound
		}
		return secret.Secret{}, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret.Secret{
		Key:  key,
		Data: data,
		Hash: hash.SliceToArray(storedHash),
	}, nil
}

// Put stores a secret in the database.
func (db DB) Put(ctx context.Context, username string, sec secret.Secret, hash [32]byte) error {
	if hash == [32]byte{} {
		// assumes no previous record of this secret exists
		_, err := db.ExecContext(ctx, addSecretQuery, username, sec.Key, sec.Data, sec.Hash[:])
		if err != nil {
			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
					return secret.ErrWrongHash
				}
			}
		}
		return err
	}

	v, err := db.ExecContext(ctx, updateSecretQuery, sec.Data, sec.Hash[:], username, sec.Key, hash[:])
	if err != nil {
		return err
	}

	rows, err := v.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return secret.ErrWrongHash
	}

	return nil
}

// Delete deletes a secret from the database.
func (db DB) Delete(ctx context.Context, username, key string, knownHash [32]byte) error {
	v, err := db.ExecContext(ctx, deleteSecretQuery, username, key, knownHash[:])
	if err != nil {
		return err
	}

	rows, err := v.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return secret.ErrWrongHash
	}

	return nil
}

// List returns a list of secrets (no data, only hashes) from the database.
func (db DB) List(ctx context.Context, username string) ([]secret.Secret, error) {
	rows, err := db.QueryContext(ctx, listHashesQuery, username)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var secrets []secret.Secret

	for rows.Next() {
		var (
			key string
			h   []byte
		)

		err := rows.Scan(&key, &h)
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, secret.Secret{
			Key:  key,
			Hash: hash.SliceToArray(h),
		})
	}

	return secrets, nil
}
