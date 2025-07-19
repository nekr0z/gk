// Package db implements the storage layer on a database.
package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// var _ secret.SecretStorage = DB{}

// DB is the database.
type DB struct {
	*sql.DB
}

// New returns a new database.
func New(dsn string) (DB, error) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return DB{}, err
	}

	_, err = database.Exec(createUserTableQuery)
	if err != nil {
		return DB{}, err
	}

	_, err = database.Exec(createSecretTableQuery)
	if err != nil {
		return DB{}, err
	}

	return DB{database}, nil
}

// Close closes the database.
func (db DB) Close() error {
	return db.DB.Close()
}
