package cli_test

import (
	"context"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/cli"
	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/internal/manager/storage/sqlite"
)

var (
	passPhrase = "test-passphrase"
)

func TestCreate_Text(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-text"

	cmd := cli.RootCmd()

	cmd.SetArgs([]string{"create", "text", secretName, "my secret note", "-d", dbFilename, "-p", passPhrase, "-m", "key1=value1"})
	cmd.Execute()

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo, err := storage.New(db, passPhrase)
	require.NoError(t, err)

	sec, err := repo.Read(context.Background(), secretName)
	require.NoError(t, err)

	v, ok := sec.GetMetadataValue("key1")
	assert.True(t, ok)
	assert.Equal(t, v, "value1")

	txt, ok := sec.Value().(*secret.Text)
	require.True(t, ok)

	require.Equal(t, "my secret note", txt.String())
}

func TestCreate_Binary(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-binary"

	bin := make([]byte, 1024)
	_, err := rand.Read(bin)
	require.NoError(t, err)

	fn := "test-file"
	filename := filepath.Join(dir, fn)

	err = os.WriteFile(filename, bin, 0644)
	require.NoError(t, err)

	cmd := cli.RootCmd()

	cmd.SetArgs([]string{"create", "binary", secretName, filename, "-d", dbFilename, "-p", passPhrase, "-m", "key2=value2"})
	cmd.Execute()

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo, err := storage.New(db, passPhrase)
	require.NoError(t, err)

	sec, err := repo.Read(context.Background(), secretName)
	require.NoError(t, err)

	v, ok := sec.GetMetadataValue("key2")
	assert.True(t, ok)
	assert.Equal(t, v, "value2")

	got, ok := sec.Value().(*secret.Binary)
	require.True(t, ok)

	require.Equal(t, got.Bytes(), bin)
}

func TestCreate_Password(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-pwd"

	user := "user@example.com"
	pwd := "my secret password"

	cmd := cli.RootCmd()

	cmd.SetArgs([]string{"create", "password", secretName, user, pwd, "-d", dbFilename, "-p", passPhrase, "-m", "key=value", "-m", "key2=value2"})
	cmd.Execute()

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo, err := storage.New(db, passPhrase)
	require.NoError(t, err)

	sec, err := repo.Read(context.Background(), secretName)
	require.NoError(t, err)

	v, ok := sec.GetMetadataValue("key")
	assert.True(t, ok)
	assert.Equal(t, v, "value")
	v, ok = sec.GetMetadataValue("key2")
	assert.True(t, ok)
	assert.Equal(t, v, "value2")

	got, ok := sec.Value().(*secret.Password)
	require.True(t, ok)

	require.Equal(t, got.Username, user)
	require.Equal(t, got.Password, pwd)
}
func TestCreate_Card(t *testing.T) {
	dir := t.TempDir()
	dbFilename := filepath.Join(dir, "test.db")

	secretName := "test-card"

	user := "JAMES BOND"
	cardNumber := "1234 5678 9012 3456"
	expiry := "01/36"
	cvv := "777"

	cmd := cli.RootCmd()

	cmd.SetArgs([]string{"create", "card", secretName, cardNumber, expiry, cvv, user, "-d", dbFilename, "-p", passPhrase, "-m", "key=value"})
	cmd.Execute()

	db, err := sqlite.New("file:" + dbFilename)
	require.NoError(t, err)

	repo, err := storage.New(db, passPhrase)
	require.NoError(t, err)

	sec, err := repo.Read(context.Background(), secretName)
	require.NoError(t, err)

	v, ok := sec.GetMetadataValue("key")
	assert.True(t, ok)
	assert.Equal(t, v, "value")

	got, ok := sec.Value().(*secret.Card)
	require.True(t, ok)

	require.Equal(t, got.Number, cardNumber)
	require.Equal(t, got.Expiry, expiry)
	require.Equal(t, got.CVV, cvv)
	require.Equal(t, got.Username, user)
}
