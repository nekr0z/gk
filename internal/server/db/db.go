// Package db implements the storage layer on a database.
package db

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var fs embed.FS

// DB is the database.
type DB struct {
	*sql.DB
}

// New returns a new database.
func New(dsn string) (DB, error) {
	if err := runMigrations(dsn); err != nil {
		return DB{}, err
	}

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return DB{}, err
	}

	return DB{database}, nil
}

// Close closes the database.
func (db DB) Close() error {
	return db.DB.Close()
}

func runMigrations(dsn string) error {
	src, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

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
