package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/manager/storage"
	secretstorage "github.com/nekr0z/gk/internal/storage"
)

func TestStorage(t *testing.T) {
	store := mockStorage{}
	passphrase := "password"
	st := storage.New(store, passphrase)

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

	err := st.Create(ctx, key, s)
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

type mockStorage map[string]secretstorage.Secret

func (m mockStorage) Get(_ context.Context, key string) (secretstorage.Secret, error) {
	value, ok := m[key]
	if !ok {
		return secretstorage.Secret{}, errors.New("not found")
	}
	return value, nil
}

func (m mockStorage) Put(_ context.Context, key string, value secretstorage.Secret) error {
	m[key] = value
	return nil
}

func (m mockStorage) Delete(_ context.Context, key string) error {
	delete(m, key)
	return nil
}
