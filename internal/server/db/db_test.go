package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/nekr0z/gk/internal/server/db"
)

var testDB db.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbName := "testdb"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(postgresContainer)
		if err != nil {
			panic(err)
		}
	}()

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testDB, err = db.New(dsn)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	err = testDB.Close()
	if err != nil {
		panic(err)
	}

	os.Exit(code)
}

func TestDB_Error(t *testing.T) {
	t.Parallel()

	d, err := db.New("invalid dsn")
	require.Error(t, err)

	assert.Panics(t, func() {
		d.Close()
	})
}
