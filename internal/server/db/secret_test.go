package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/server/secret"
)

func TestPut(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("hash mismatch on empty", func(t *testing.T) {
		err := testDB.Put(ctx, testUsername, secret.Secret{
			Key:  "key1",
			Data: []byte("data1"),
			Hash: [32]byte{'h', '1'},
		}, [32]byte{'h', '0'})
		require.Error(t, err)
		assert.True(t, errors.Is(err, secret.ErrWrongHash), "ErrWrongHash is expected")
	})

	t.Run("success on empty", func(t *testing.T) {
		err := testDB.Put(ctx, testUsername, secret.Secret{
			Key:  "key1",
			Data: []byte("data1"),
			Hash: [32]byte{'h', '1'},
		}, [32]byte{})
		require.NoError(t, err)
	})

	t.Run("hash mismatch on existing", func(t *testing.T) {
		err := testDB.Put(ctx, testUsername, secret.Secret{
			Key:  "key1",
			Data: []byte("data2"),
			Hash: [32]byte{'h', '2'},
		}, [32]byte{'h', '0'})
		require.Error(t, err)
		assert.True(t, errors.Is(err, secret.ErrWrongHash), "ErrWrongHash is expected")
	})

	t.Run("null hash on existing", func(t *testing.T) {
		err := testDB.Put(ctx, testUsername, secret.Secret{
			Key:  "key1",
			Data: []byte("data2"),
			Hash: [32]byte{'h', '2'},
		}, [32]byte{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, secret.ErrWrongHash), "ErrWrongHash is expected")
	})

	t.Run("successul update", func(t *testing.T) {
		err := testDB.Put(ctx, testUsername, secret.Secret{
			Key:  "key1",
			Data: []byte("data2"),
			Hash: [32]byte{'h', '2'},
		}, [32]byte{'h', '1'})
		require.NoError(t, err)
	})
}

func TestGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	err := testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key2",
		Data: []byte("data2"),
		Hash: [32]byte{'h', '2'},
	}, [32]byte{})
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		_, err := testDB.Get(ctx, testUsername, "key0")
		require.Error(t, err)
		assert.True(t, errors.Is(err, secret.ErrNotFound), "ErrNotFound is expected")
	})

	t.Run("success", func(t *testing.T) {
		s, err := testDB.Get(ctx, testUsername, "key2")
		require.NoError(t, err)
		assert.Equal(t, secret.Secret{
			Key:  "key2",
			Data: []byte("data2"),
			Hash: [32]byte{'h', '2'},
		}, s)
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	err := testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key3",
		Data: []byte("data3"),
		Hash: [32]byte{'h', '3'},
	}, [32]byte{})
	require.NoError(t, err)

	t.Run("hash mismatch", func(t *testing.T) {
		err := testDB.Delete(ctx, testUsername, "key3", [32]byte{'h', '1'})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, secret.ErrWrongHash), "ErrWrongHash is expected")
	})

	t.Run("success", func(t *testing.T) {
		err := testDB.Delete(ctx, testUsername, "key3", [32]byte{'h', '3'})
		assert.NoError(t, err)
	})
}

func TestList(t *testing.T) {
	testUsername := "listuser"
	t.Parallel()
	ctx := context.Background()

	err := testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key1",
		Data: []byte("data1"),
		Hash: [32]byte{'h', '1'},
	}, [32]byte{})
	require.NoError(t, err)

	err = testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key2",
		Data: []byte("data2"),
		Hash: [32]byte{'h', '2'},
	}, [32]byte{})
	require.NoError(t, err)

	err = testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key3",
		Data: []byte("data3"),
		Hash: [32]byte{'h', '3'},
	}, [32]byte{})
	require.NoError(t, err)

	err = testDB.Put(ctx, testUsername, secret.Secret{
		Key:  "key4",
		Data: []byte("data4"),
		Hash: [32]byte{'h', '4'},
	}, [32]byte{})
	require.NoError(t, err)

	s, err := testDB.List(ctx, testUsername)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(s))
	assert.Contains(t, s, secret.Secret{
		Key:  "key1",
		Hash: [32]byte{'h', '1'},
	})
	assert.Contains(t, s, secret.Secret{
		Key:  "key2",
		Hash: [32]byte{'h', '2'},
	})
	assert.Contains(t, s, secret.Secret{
		Key:  "key3",
		Hash: [32]byte{'h', '3'},
	})
	assert.Contains(t, s, secret.Secret{
		Key:  "key4",
		Hash: [32]byte{'h', '4'},
	})
}
