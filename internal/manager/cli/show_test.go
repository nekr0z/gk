package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/internal/manager/storage/sqlite"
)

func TestShow_Text(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-text"

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo := storage.New(db, passPhrase)

	sec := secret.NewText("my secret note")
	sec.SetMetadataValue("key1", "value1")
	err = repo.Create(context.Background(), secretName, sec)
	require.NoError(t, err)

	cmd := rootCmd()

	b := &bytes.Buffer{}
	cmd.SetOut(b)

	cmd.SetArgs([]string{"show", secretName, "-d", dbFilename, "-p", passPhrase})
	cmd.Execute()

	out, err := io.ReadAll(b)
	require.NoError(t, err)

	assert.Contains(t, string(out), "my secret note")
	assert.Contains(t, string(out), "key1: value1")

	b = &bytes.Buffer{}

	cmd = rootCmd()
	cmd.SetOut(b)

	filename := filepath.Join(dir, "test.txt")

	cmd.SetArgs([]string{"show", secretName, "-d", dbFilename, "-p", passPhrase, "-t", filename})
	cmd.Execute()

	out, err = io.ReadAll(b)
	require.NoError(t, err)

	assert.Contains(t, string(out), "my secret note")

	assert.FileExists(t, filename)

	f, err := os.Open(filename)
	require.NoError(t, err)
	defer f.Close()

	txt, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Equal(t, "my secret note", string(txt))
}

func TestShow_Binary(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-bin"

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo := storage.New(db, passPhrase)

	bin := make([]byte, 1024)
	_, err = rand.Read(bin)
	require.NoError(t, err)

	sec := secret.NewBinary(bin)
	err = repo.Create(context.Background(), secretName, sec)
	require.NoError(t, err)

	cmd := rootCmd()

	b := &bytes.Buffer{}
	cmd.SetOut(b)

	cmd.SetArgs([]string{"show", secretName, "-d", dbFilename, "-p", passPhrase})
	cmd.Execute()

	out, err := io.ReadAll(b)
	require.NoError(t, err)

	assert.Contains(t, string(out), "BINARY")

	b = &bytes.Buffer{}

	cmd = rootCmd()
	cmd.SetOut(b)

	filename := filepath.Join(dir, "test.bin")

	cmd.SetArgs([]string{"show", secretName, "-d", dbFilename, "-p", passPhrase, "-t", filename})
	cmd.Execute()

	out, err = io.ReadAll(b)
	require.NoError(t, err)

	assert.Contains(t, string(out), "*BINARY DATA*")

	assert.FileExists(t, filename)

	f, err := os.Open(filename)
	require.NoError(t, err)
	defer f.Close()

	bb, err := io.ReadAll(f)
	require.NoError(t, err)

	assert.Equal(t, bin, bb)
}
