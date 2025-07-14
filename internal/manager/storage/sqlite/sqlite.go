// Package sqlite provides an SQLite3 implementation of the storage.Storage
// interface.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
)

const (
	tableName = "secrets"

	createTableQuery = `
	CREATE TABLE IF NOT EXISTS ` + tableName + `(
		id TEXT PRIMARY KEY,
		encrypted_payload BLOB NOT NULL,
		payload_hash BLOB NOT NULL,
		server_hash BLOB NOT NULL
	)`
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

// Storage implements the storage.Storage interface using an SQL database.
type Storage struct {
	db *sql.DB
}

// New creates a new SQL storage instance using the provided DSN.
// Secrets table is created if one doesn't exist.
func New(dsn string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTableIfNotExists(db); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &Storage{db: db}, nil
}

func createTableIfNotExists(db *sql.DB) error {
	_, err := db.Exec(createTableQuery)
	return err
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
			return storage.StoredSecret{}, fmt.Errorf("secret not found")
		}
		return storage.StoredSecret{}, fmt.Errorf("failed to get secret: %w", err)
	}

	serverHashArr := toHashArray(serverHash)

	return storage.StoredSecret{
		EncryptedPayload: crypt.Data{
			Data: encryptedPayload,
			Hash: toHashArray(payloadHash),
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

func toHashArray(b []byte) [32]byte {
	if len(b) != 32 {
		panic("invalid hash length")
	}

	var arr [32]byte
	copy(arr[:], b)

	return arr
}
