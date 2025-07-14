package sqlite

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	db, err := New("file::memory:?cache=shared")
	require.NoError(t, err, "Failed to create database")
	defer db.Close()

	key := "test-key"
	payload := []byte("encrypted-data")
	hash := sha256.Sum256(payload)
	secret := storage.StoredSecret{
		EncryptedPayload: crypt.Data{
			Data: payload,
			Hash: hash,
		},
		LastKnownServerHash: [32]byte{1, 2, 3},
	}

	t.Run("PutNewSecret", func(t *testing.T) {
		err := db.Put(ctx, key, secret)
		assert.NoError(t, err, "Put failed")
	})

	t.Run("GetExistingSecret", func(t *testing.T) {
		got, err := db.Get(ctx, key)
		require.NoError(t, err, "Get failed")

		assert.Equal(t, secret.EncryptedPayload.Data, got.EncryptedPayload.Data, "Payload data mismatch")
		assert.Equal(t, secret.EncryptedPayload.Hash, got.EncryptedPayload.Hash, "Payload hash mismatch")
		assert.Equal(t, secret.LastKnownServerHash, got.LastKnownServerHash, "Server hash mismatch")
	})

	t.Run("DeleteSecret", func(t *testing.T) {
		err := db.Delete(ctx, key)
		assert.NoError(t, err, "Delete failed")

		_, err = db.Get(ctx, key)
		assert.Error(t, err, "Expected error after deletion")
	})
}
