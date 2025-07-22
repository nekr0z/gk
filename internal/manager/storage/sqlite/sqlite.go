// Package sqlite provides an SQLite3 implementation of the storage.Storage
// interface.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"

	"github.com/nekr0z/gk/internal/hash"
	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
)

const (
	tableName = "secrets"

	selectQuery = `SELECT encrypted_payload, payload_hash, server_hash FROM ` + tableName + ` WHERE id = ?`
	insertQuery = `INSERT INTO ` + tableName + `
	(id, encrypted_payload, payload_hash, server_hash)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		encrypted_payload = excluded.encrypted_payload,
		payload_hash = excluded.payload_hash,
		server_hash = excluded.server_hash`
	deleteQuery = `DELETE FROM ` + tableName + ` WHERE id = ?`
)

//go:embed migrations/*.sql
var fs embed.FS

var _ storage.Storage = (*Storage)(nil)

// Storage implements the storage.Storage interface using an SQL database.
type Storage struct {
	db *sql.DB
}

// New creates a new SQL storage instance using the provided DSN.
// Secrets table is created if one doesn't exist.
func New(dsn string) (*Storage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Storage{db: db}, nil
}

func runMigrations(dsn string) error {
	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	dsn = strings.TrimPrefix(dsn, "file:")
	dsn = fmt.Sprintf("sqlite3://%s", dsn)

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// Close closes the database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// Get retrieves a secret by its key from the database.
func (s *Storage) Get(ctx context.Context, key string) (storage.StoredSecret, error) {
	row := s.db.QueryRowContext(ctx, selectQuery, key)

	var (
		encryptedPayload []byte
		payloadHash      []byte
		serverHash       []byte
	)

	if err := row.Scan(&encryptedPayload, &payloadHash, &serverHash); err != nil {
		if err == sql.ErrNoRows {
			return storage.StoredSecret{}, storage.ErrNotFound
		}
		return storage.StoredSecret{}, fmt.Errorf("failed to get secret: %w", err)
	}

	serverHashArr := hash.SliceToArray(serverHash)

	return storage.StoredSecret{
		EncryptedPayload: crypt.Data{
			Data: encryptedPayload,
			Hash: hash.SliceToArray(payloadHash),
		},
		LastKnownServerHash: serverHashArr,
	}, nil
}

// Put stores a secret in the database with the given key.
func (s *Storage) Put(ctx context.Context, key string, secret storage.StoredSecret) error {
	_, err := s.db.ExecContext(ctx, insertQuery,
		key,
		secret.EncryptedPayload.Data,
		secret.EncryptedPayload.Hash[:],
		secret.LastKnownServerHash[:],
	)

	return err
}

// Delete removes a secret from the database by its key.
func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.db.ExecContext(ctx, deleteQuery, key)

	return err
}

// List retrieves a list of all secrets from the database.
func (s *Storage) List(ctx context.Context) (map[string]storage.ListedSecret, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, payload_hash, server_hash FROM "+tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	secrets := make(map[string]storage.ListedSecret)
	for rows.Next() {
		var (
			id          string
			payloadHash []byte
			serverHash  []byte
		)
		if err = rows.Scan(&id, &payloadHash, &serverHash); err != nil {
			return nil, err
		}
		secrets[id] = storage.ListedSecret{
			Hash:                hash.SliceToArray(payloadHash),
			LastKnownServerHash: hash.SliceToArray(serverHash),
		}
	}

	return secrets, nil
}
