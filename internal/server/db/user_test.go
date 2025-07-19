package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/gk/internal/server/user"
)

var (
	testUsername = "testuser"
	testPassword = []byte("testpassword")
)

func TestAddUser(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		err := testDB.AddUser(ctx, &user.User{
			Username: testUsername,
			Password: testPassword,
		})
		assert.NoError(t, err)
	})

	t.Run("duplicate username", func(t *testing.T) {
		err := testDB.AddUser(ctx, &user.User{
			Username: testUsername,
			Password: testPassword,
		})
		assert.Error(t, err)
		assert.True(t, errors.Is(err, user.ErrAlreadyExists))
	})
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		u, err := testDB.GetUser(ctx, testUsername)
		assert.NoError(t, err)
		assert.Equal(t, testPassword, u.Password)
		assert.Equal(t, testUsername, u.Username)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		_, err := testDB.GetUser(ctx, "notfound")
		assert.Error(t, err)
	})
}
