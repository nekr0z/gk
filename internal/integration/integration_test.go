package integration

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	gk "github.com/nekr0z/gk/internal/manager/cli"
	server "github.com/nekr0z/gk/internal/server/cli"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// set up Postgres
	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp")),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(postgresContainer)
		require.NoError(t, err)
	})

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// start up the server
	serverChan := make(chan struct{})
	t.Cleanup(func() {
		<-serverChan
	})

	serverCtx, serverCancel := context.WithCancel(ctx)
	t.Cleanup(serverCancel)

	serverCmd := server.RootCmd()
	serverCmd.SetArgs([]string{"--dsn", dsn, "--address", ":"})

	pr, pw := io.Pipe()
	serverCmd.SetOut(pw)

	go func() {
		serverCmd.ExecuteContext(serverCtx)
		close(serverChan)
	}()

	var serverAddr string

	sc := bufio.NewScanner(pr)

	for sc.Scan() {
		line := sc.Text()
		serverAddr = strings.TrimPrefix(line, "Running server on ")
		break
	}

	go io.Copy(io.Discard, pr)

	require.NotEmpty(t, serverAddr)

	// set up two clients
	tempDir := t.TempDir()
	db1 := filepath.Join(tempDir, "db1")
	db2 := filepath.Join(tempDir, "db2")
	passphrase := "my very secret passphrase"

	t.Run("add some secrets on first client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"create", "text", "secret-note", "This note is very secret!", "-d", db1, "-p", passphrase})

		err := cmd.Execute()
		assert.NoError(t, err)

		cmd = gk.RootCmd()
		cmd.SetArgs([]string{"create", "password", "example-password", "user@example.com", "monkey123", "-d", db1, "-p", passphrase, "-m", "url=example.com"})

		err = cmd.Execute()
		assert.NoError(t, err)
	})

	username := "user"
	password := "password"

	t.Run("register user", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"signup", "-i", "-s", serverAddr, "-u", username, "-w", password})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("first sync with server", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db1})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("add a secret on the second client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"create", "text", "another-note", "This note is even more secret than the other one!", "-d", db2, "-p", passphrase})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("sync the second client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db2})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("check that the second client has the secret", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"show", "secret-note", "-d", db2, "-p", passphrase})
		b := &bytes.Buffer{}
		cmd.SetOut(b)

		err := cmd.Execute()
		assert.NoError(t, err)

		out, err := io.ReadAll(b)
		require.NoError(t, err)

		assert.Contains(t, string(out), "This note is very secret!")
	})

	t.Run("delete and re-create (i.e. update) the secret on the second client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"delete", "secret-note", "-d", db2, "-p", passphrase})

		err := cmd.Execute()
		assert.NoError(t, err)

		cmd = gk.RootCmd()
		cmd.SetArgs([]string{"create", "text", "secret-note", "This note is very secret again, but with a different content.", "-d", db2, "-p", passphrase})

		err = cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("sync the second client again", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db2})

		err := cmd.Execute()
		assert.Error(t, err, "conflict expected")
	})

	t.Run("sync the second client with resolver", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db2, "-g", "local"})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("sync the first client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db1})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("check that the first client has the secret", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"show", "secret-note", "-d", db1, "-p", passphrase})
		b := &bytes.Buffer{}
		cmd.SetOut(b)

		err := cmd.Execute()
		assert.NoError(t, err)

		out, err := io.ReadAll(b)
		require.NoError(t, err)

		assert.Contains(t, string(out), "This note is very secret again")
	})

	t.Run("re-register user", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"signup", "-i", "-s", serverAddr, "-u", username, "-w", "someotherpassword"})

		err := cmd.Execute()
		assert.Error(t, err, "already registered")
	})

	t.Run("delete the secret on the first client", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"delete", "secret-note", "-d", db1, "-p", passphrase})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("check that the first client no longer has the secret", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"show", "secret-note", "-d", db1, "-p", passphrase})

		err := cmd.Execute()
		assert.Error(t, err, "should not be found")
	})

	t.Run("sync the first client after deletion", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db1})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("sync the second client after deletion", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"sync", "-i", "-s", serverAddr, "-u", username, "-w", password, "-d", db2})

		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("check that the second client no longer has the secret", func(t *testing.T) {
		cmd := gk.RootCmd()
		cmd.SetArgs([]string{"show", "secret-note", "-d", db2, "-p", passphrase})

		err := cmd.Execute()
		assert.Error(t, err, "should not be found")
	})
}
