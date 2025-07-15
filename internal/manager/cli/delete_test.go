package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/internal/manager/storage/sqlite"
)

func TestDelete(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test"

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo := storage.New(db, passPhrase)

	sec := secret.NewText("my secret note")
	err = repo.Create(context.Background(), secretName, sec)
	require.NoError(t, err)

	cmd := rootCmd()

	cmd.SetArgs([]string{"delete", secretName, "-d", dbFilename, "-p", passPhrase})
	cmd.Execute()

	sec, err = repo.Read(context.Background(), secretName)
	assert.Error(t, err)
	assert.Nil(t, sec.Metadata())
	assert.Nil(t, sec.Value())
}
