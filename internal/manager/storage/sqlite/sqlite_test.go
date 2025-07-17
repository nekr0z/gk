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

	key2 := "test-key2"
	payload = []byte("some-other-encrypted-data")
	hash = sha256.Sum256(payload)
	secret2 := storage.StoredSecret{
		EncryptedPayload: crypt.Data{
			Data: payload,
			Hash: hash,
		},
		LastKnownServerHash: [32]byte{1, 2, 3, 4},
	}

	t.Run("PutNewSecret", func(t *testing.T) {
		err := db.Put(ctx, key, secret)
		assert.NoError(t, err, "Put failed")

		err = db.Put(ctx, key2, secret2)
		assert.NoError(t, err, "Put failed")
	})

	t.Run("ListSecrets", func(t *testing.T) {
		secrets, err := db.List(ctx)
		assert.NoError(t, err, "List failed")
		assert.Len(t, secrets, 2, "Expected 2 secrets")

		s1 := storage.ListedSecret{
			Hash:                secret.EncryptedPayload.Hash,
			LastKnownServerHash: secret.LastKnownServerHash,
		}
		s2 := storage.ListedSecret{
			Hash:                secret2.EncryptedPayload.Hash,
			LastKnownServerHash: secret2.LastKnownServerHash,
		}
		assert.Equal(t, secrets[key], s1, "Expected secret 1 to be in the list")
		assert.Equal(t, secrets[key2], s2, "Expected secret 2 to be in the list")
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

func TestToHashArray(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var expected [32]byte
		result := toHashArray(nil)
		assert.Equal(t, expected, result)
	})

	t.Run("valid 32 bytes", func(t *testing.T) {
		b := make([]byte, 32)
		for i := range b {
			b[i] = byte(i)
		}
		var expected [32]byte
		copy(expected[:], b)
		result := toHashArray(b)
		assert.Equal(t, expected, result)
	})

	t.Run("invalid length", func(t *testing.T) {
		b := make([]byte, 31)
		assert.Panics(t, func() {
			toHashArray(b)
		})
	})
}
