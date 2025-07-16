package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
)

var (
	testKey        = "test"
	testPassphrase = "password"

	emptyHash = [32]byte{}
	hash1     = [32]byte{1, 2, 3}
	hash2     = [32]byte{1, 2, 3, 4}
	hash3     = [32]byte{1, 2, 3, 4, 5}
	hash4     = [32]byte{1, 2, 3, 4, 5, 6}

	payload1 = crypt.Data{
		Data: []byte("test"),
		Hash: hash1,
	}
	payload2 = crypt.Data{
		Data: []byte("test2"),
		Hash: hash2,
	}
	payload3 = crypt.Data{
		Data: []byte("test3"),
		Hash: hash3,
	}
	payload4 = crypt.Data{
		Data: []byte("test3"),
		Hash: hash4,
	}
)

func TestSync(t *testing.T) {
	ctx := context.Background()

	rem := storage.NewMockRemote(t)
	loc := storage.NewMockStorage(t)

	repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem))
	require.NoError(t, err)

	check := func(t *testing.T) {
		t.Helper()

		err := repo.Sync(ctx, testKey)
		require.NoError(t, err)

		loc.AssertExpectations(t)
		rem.AssertExpectations(t)
	}

	t.Run("created local to remote", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload: payload1,
		}, nil).Twice()

		mock.InOrder(
			rem.On("Get", mock.Anything, testKey).Return(crypt.Data{}, storage.ErrNotFound).Once(),
			rem.On("Put", mock.Anything, testKey, payload1, emptyHash).Return(nil),
			loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
				EncryptedPayload:    payload1,
				LastKnownServerHash: hash1,
			}).Return(nil).Once(),
		)

		check(t)
	})

	t.Run("created remote to local", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{}, storage.ErrNotFound).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload1, nil).Once()
		loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
			EncryptedPayload:    payload1,
			LastKnownServerHash: hash1,
		}).Return(nil).Once()

		check(t)
	})

	t.Run("synced", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload1,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload1, nil).Once()

		check(t)
	})

	t.Run("unsynced but equal", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()
		loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash2,
		}).Return(nil).Once()

		check(t)
	})

	t.Run("updated local to remote", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash1,
		}, nil).Twice()
		rem.On("Get", mock.Anything, testKey).Return(payload1, nil).Once()
		loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash2,
		}).Return(nil).Once()
		rem.On("Put", mock.Anything, testKey, payload2, hash1).Return(nil).Once()

		check(t)
	})

	t.Run("updated remote to local", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload1,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()
		loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash2,
		}).Return(nil).Once()

		check(t)
	})

	t.Run("deleted locally", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			LastKnownServerHash: hash1,
		}, nil).Twice()
		rem.On("Get", mock.Anything, testKey).Return(payload1, nil).Once()
		loc.On("Delete", mock.Anything, testKey).Return(nil).Once()
		rem.On("Delete", mock.Anything, testKey, hash1).Return(nil).Once()

		check(t)
	})

	t.Run("deleted remotely", func(t *testing.T) {
		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload1,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(crypt.Data{}, storage.ErrNotFound).Once()
		loc.On("Delete", mock.Anything, testKey).Return(nil).Once()

		check(t)
	})
}

func TestSync_Conflict(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rem := storage.NewMockRemote(t)
	loc := storage.NewMockStorage(t)

	t.Run("no resolver", func(t *testing.T) {
		repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem))
		require.NoError(t, err)

		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload3,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()

		err = repo.Sync(ctx, testKey)
		require.Error(t, err)
	})

	t.Run("remote wins", func(t *testing.T) {
		repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem), storage.UseResolver(
			func(ctx context.Context, local, remote crypt.Data) (crypt.Data, error) {
				return remote, nil
			}))
		require.NoError(t, err)

		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload3,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()
		loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
			EncryptedPayload:    payload2,
			LastKnownServerHash: hash2,
		}).Return(nil).Once()

		err = repo.Sync(ctx, testKey)
		require.NoError(t, err)
	})

	t.Run("local wins", func(t *testing.T) {
		repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem), storage.UseResolver(
			func(ctx context.Context, local, remote crypt.Data) (crypt.Data, error) {
				return local, nil
			}))
		require.NoError(t, err)

		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload3,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()
		rem.On("Put", mock.Anything, testKey, payload3, hash2).Return(nil).Once()
		mock.InOrder(
			loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
				EncryptedPayload:    payload3,
				LastKnownServerHash: hash2,
			}).Return(nil).Once(),
			loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
				EncryptedPayload:    payload3,
				LastKnownServerHash: hash2,
			}, nil).Once(),
			loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
				EncryptedPayload:    payload3,
				LastKnownServerHash: hash3,
			}).Return(nil).Once(),
		)

		err = repo.Sync(ctx, testKey)
		require.NoError(t, err)
	})

	t.Run("new state", func(t *testing.T) {
		repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem), storage.UseResolver(
			func(ctx context.Context, local, remote crypt.Data) (crypt.Data, error) {
				return payload4, nil
			}))
		require.NoError(t, err)

		loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
			EncryptedPayload:    payload3,
			LastKnownServerHash: hash1,
		}, nil).Once()
		rem.On("Get", mock.Anything, testKey).Return(payload2, nil).Once()
		rem.On("Put", mock.Anything, testKey, payload4, hash2).Return(nil).Once()
		mock.InOrder(
			loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
				EncryptedPayload:    payload4,
				LastKnownServerHash: hash2,
			}).Return(nil).Once(),
			loc.On("Get", mock.Anything, testKey).Return(storage.StoredSecret{
				EncryptedPayload:    payload4,
				LastKnownServerHash: hash2,
			}, nil).Once(),
			loc.On("Put", mock.Anything, testKey, storage.StoredSecret{
				EncryptedPayload:    payload4,
				LastKnownServerHash: hash4,
			}).Return(nil).Once(),
		)

		err = repo.Sync(ctx, testKey)
		require.NoError(t, err)
	})
}

func TestSyncAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	rem := storage.NewMockRemote(t)
	loc := storage.NewMockStorage(t)

	repo, err := storage.New(loc, testPassphrase, storage.UseRemote(rem))
	require.NoError(t, err)

	loc.On("List", mock.Anything).Return(map[string]storage.ListedSecret{
		"key": {
			Hash:                hash1,
			LastKnownServerHash: hash1,
		},
		"key2": {
			Hash: hash2,
		},
		"key3": {
			Hash:                hash3,
			LastKnownServerHash: hash2,
		},
	}, nil).Once()
	rem.On("List", mock.Anything).Return([]storage.RemoteListedSecret{
		{Key: "key", Hash: hash1},
		{Key: "key3", Hash: hash2},
		{Key: "key4", Hash: hash4},
	}, nil).Once()

	loc.On("Get", mock.Anything, "key2").Return(storage.StoredSecret{
		EncryptedPayload: payload2,
	}, nil).Twice()
	rem.On("Get", mock.Anything, "key2").Return(crypt.Data{}, storage.ErrNotFound).Once()
	rem.On("Put", mock.Anything, "key2", payload2, emptyHash).Return(nil).Once()
	loc.On("Put", mock.Anything, "key2", storage.StoredSecret{
		EncryptedPayload:    payload2,
		LastKnownServerHash: hash2,
	}).Return(nil).Once()

	loc.On("Get", mock.Anything, "key3").Return(storage.StoredSecret{
		EncryptedPayload:    payload3,
		LastKnownServerHash: hash2,
	}, nil).Twice()
	rem.On("Get", mock.Anything, "key3").Return(payload2, nil).Once()
	loc.On("Put", mock.Anything, "key3", storage.StoredSecret{
		EncryptedPayload:    payload3,
		LastKnownServerHash: hash3,
	}).Return(nil).Once()
	rem.On("Put", mock.Anything, "key3", payload3, hash2).Return(nil).Once()

	loc.On("Get", mock.Anything, "key4").Return(storage.StoredSecret{}, storage.ErrNotFound).Once()
	rem.On("Get", mock.Anything, "key4").Return(payload4, nil).Once()
	loc.On("Put", mock.Anything, "key4", storage.StoredSecret{
		EncryptedPayload:    payload4,
		LastKnownServerHash: hash4,
	}).Return(nil).Once()

	err = repo.SyncAll(ctx)
	require.NoError(t, err)

	loc.AssertExpectations(t)
	rem.AssertExpectations(t)
}
