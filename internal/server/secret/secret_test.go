package secret

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		expected := Secret{Key: "test", Data: []byte("data")}
		mockStorage.On("Get", mock.Anything, "user1", "test").Return(expected, nil)

		result, err := svc.GetSecret(context.Background(), "user1", "test")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		_, err := svc.GetSecret(context.Background(), "", "test")

		require.ErrorIs(t, err, ErrNoUser)
		mockStorage.AssertNotCalled(t, "Get")
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		mockStorage.On("Get", mock.Anything, "user1", "test").Return(Secret{}, ErrNotFound)

		_, err := svc.GetSecret(context.Background(), "user1", "test")

		require.ErrorIs(t, err, ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_PutSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		secret := Secret{Key: "test", Data: []byte("data")}
		var hash [32]byte
		mockStorage.On("Put", mock.Anything, "user1", secret, hash).Return(nil)

		err := svc.PutSecret(context.Background(), "user1", secret, hash)

		require.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		err := svc.PutSecret(context.Background(), "", Secret{}, [32]byte{})

		require.ErrorIs(t, err, ErrNoUser)
		mockStorage.AssertNotCalled(t, "Put")
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		secret := Secret{Key: "test", Data: []byte("data")}
		var hash [32]byte
		mockStorage.On("Put", mock.Anything, "user1", secret, hash).Return(ErrWrongHash)

		err := svc.PutSecret(context.Background(), "user1", secret, hash)

		require.ErrorIs(t, err, ErrWrongHash)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_DeleteSecret(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		var hash [32]byte
		mockStorage.On("Delete", mock.Anything, "user1", "test", hash).Return(nil)

		err := svc.DeleteSecret(context.Background(), "user1", "test", hash)

		require.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		err := svc.DeleteSecret(context.Background(), "", "test", [32]byte{})

		require.ErrorIs(t, err, ErrNoUser)
		mockStorage.AssertNotCalled(t, "Delete")
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		var hash [32]byte
		mockStorage.On("Delete", mock.Anything, "user1", "test", hash).Return(ErrNotFound)

		err := svc.DeleteSecret(context.Background(), "user1", "test", hash)

		require.ErrorIs(t, err, ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}

func TestService_ListSecrets(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		expected := []Secret{
			{Key: "test1", Hash: [32]byte{}},
			{Key: "test2", Hash: [32]byte{}},
		}
		mockStorage.On("List", mock.Anything, "user1").Return(expected, nil)

		result, err := svc.ListSecrets(context.Background(), "user1")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mockStorage.AssertExpectations(t)
	})

	t.Run("empty username", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		_, err := svc.ListSecrets(context.Background(), "")

		require.ErrorIs(t, err, ErrNoUser)
		mockStorage.AssertNotCalled(t, "List")
	})

	t.Run("storage error", func(t *testing.T) {
		t.Parallel()
		mockStorage := new(MockSecretStorage)
		svc := NewService(mockStorage)

		mockStorage.On("List", mock.Anything, "user1").Return(nil, ErrNotFound)

		_, err := svc.ListSecrets(context.Background(), "user1")

		require.ErrorIs(t, err, ErrNotFound)
		mockStorage.AssertExpectations(t)
	})
}
