package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/manager/storage"
)

func TestStorage(t *testing.T) {
	store := mockStorage{}
	passphrase := "password"
	st, err := storage.New(store, passphrase)
	require.NoError(t, err)

	note := "my note"
	s := secret.NewText(note)
	meta := map[string]string{
		"key":  "value",
		"key2": "value2",
		"key3": "value3",
	}
	s.SetMetadata(meta)
	key := "my key"

	ctx := context.Background()

	err = st.Create(ctx, key, s)
	require.NoError(t, err)

	bb := s.Marshal()
	d := store[key]
	assert.NotEqual(t, d.EncryptedPayload.Data, bb, "should be encrypted")

	got, err := st.Read(ctx, key)
	require.NoError(t, err)

	v := got.Value()
	txt, ok := v.(*secret.Text)
	require.True(t, ok)

	assert.Equal(t, note, txt.String())
	assert.Equal(t, meta, got.Metadata())

	err = st.Delete(ctx, key)
	require.NoError(t, err)

	_, err = st.Read(ctx, key)
	assert.Error(t, err)
}

func TestNew_Error(t *testing.T) {
	t.Parallel()

	_, err := storage.New(nil, testPassphrase)
	require.Error(t, err, "expected error on nil storage")
}

func TestCreate_Error(t *testing.T) {
	t.Parallel()

	st := storage.NewMockStorage(t)
	r, err := storage.New(st, testPassphrase)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = r.Create(ctx, "test", secret.NewText("test"))
	assert.Error(t, err, "context canceled")
}

func TestRead_Error(t *testing.T) {
	t.Parallel()

	t.Run("context canceled", func(t *testing.T) {
		t.Parallel()
		st := storage.NewMockStorage(t)
		r, err := storage.New(st, testPassphrase)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = r.Read(ctx, "test")
		assert.Error(t, err, "context canceled")
	})

	t.Run("deleted locally", func(t *testing.T) {
		t.Parallel()
		st := storage.NewMockStorage(t)
		r, err := storage.New(st, testPassphrase)
		require.NoError(t, err)

		st.EXPECT().Get(mock.Anything, "test").Return(storage.StoredSecret{
			EncryptedPayload:    crypt.Data{},
			LastKnownServerHash: [32]byte{'h'},
		}, nil)
		_, err = r.Read(context.Background(), "test")
		assert.Error(t, err, "not found")
	})

	t.Run("failed to decrypt", func(t *testing.T) {
		t.Parallel()
		st := storage.NewMockStorage(t)
		r, err := storage.New(st, testPassphrase)
		require.NoError(t, err)

		st.EXPECT().Get(mock.Anything, "test").Return(storage.StoredSecret{
			EncryptedPayload: crypt.Data{
				Data: []byte("test"),
				Hash: [32]byte{'h'},
			},
			LastKnownServerHash: [32]byte{},
		}, nil)
		_, err = r.Read(context.Background(), "test")
		assert.Error(t, err, "failed to decrypt")
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	st := storage.NewMockStorage(t)
	r, err := storage.New(st, testPassphrase)
	require.NoError(t, err)

	st.EXPECT().Get(mock.Anything, "test").Return(storage.StoredSecret{}, storage.ErrNotFound)

	err = r.Delete(context.Background(), "test")
	assert.NoError(t, err)
}

func TestDelete_Error(t *testing.T) {
	t.Parallel()

	st := storage.NewMockStorage(t)
	r, err := storage.New(st, testPassphrase)
	require.NoError(t, err)

	t.Run("context canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = r.Delete(ctx, "test")
		assert.Error(t, err, "context canceled")
	})

	t.Run("storage error", func(t *testing.T) {
		st.EXPECT().Get(mock.Anything, "test").Return(storage.StoredSecret{}, errors.New("storage error"))

		err = r.Delete(context.Background(), "test")
		assert.Error(t, err, "storage error")
	})
}

type mockStorage map[string]storage.StoredSecret

func (m mockStorage) Get(_ context.Context, key string) (storage.StoredSecret, error) {
	value, ok := m[key]
	if !ok {
		return storage.StoredSecret{}, errors.New("not found")
	}
	return value, nil
}

func (m mockStorage) Put(_ context.Context, key string, value storage.StoredSecret) error {
	m[key] = value
	return nil
}

func (m mockStorage) Delete(_ context.Context, key string) error {
	delete(m, key)
	return nil
}

func (m mockStorage) List(_ context.Context) (map[string]storage.ListedSecret, error) {
	return nil, nil
}
