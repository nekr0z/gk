package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUserService(t *testing.T) {
	t.Parallel()

	t.Run("default token signing key", func(t *testing.T) {
		t.Parallel()

		mockStorage := NewMockUserStorage(t)
		service, err := NewUserService(mockStorage)
		require.NoError(t, err)
		assert.NotNil(t, service.tokenSigningKey)
		assert.Len(t, service.tokenSigningKey, tokenDefaultLen)
	})

	t.Run("custom token signing key", func(t *testing.T) {
		t.Parallel()

		mockStorage := NewMockUserStorage(t)
		customKey := []byte("custom-signing-key-12345")
		service, err := NewUserService(mockStorage, WithTokenSigningKey(customKey))
		require.NoError(t, err)
		assert.Equal(t, customKey, service.tokenSigningKey)
	})
}

func TestUserService_Register(t *testing.T) {

	ctx := context.Background()
	mockStorage := NewMockUserStorage(t)
	service, err := NewUserService(mockStorage)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		mockStorage.On("AddUser", ctx, mock.AnythingOfType("*user.User")).
			Return(nil).
			Once()

		err := service.Register(ctx, "testuser", "password123")
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockStorage.On("AddUser", ctx, mock.AnythingOfType("*user.User")).
			Return(ErrAlreadyExists).
			Once()

		err := service.Register(ctx, "existinguser", "password123")
		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockStorage.AssertExpectations(t)
	})

	t.Run("password hashing", func(t *testing.T) {
		mockStorage.On("AddUser", ctx, mock.MatchedBy(func(u *User) bool {
			return bcrypt.CompareHashAndPassword(u.Password, []byte("password123")) == nil
		})).Return(nil).Once()

		err := service.Register(ctx, "hashuser", "password123")
		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
	})
}

func TestUserService_CreateToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := NewMockUserStorage(t)
	fixedKey := []byte("fixed-signing-key-for-tests")
	service, err := NewUserService(mockStorage, WithTokenSigningKey(fixedKey))
	require.NoError(t, err)

	validUser := &User{
		Username: "validuser",
		Password: mustHashPassword("correctpassword"),
	}

	testCases := []struct {
		name        string
		username    string
		password    string
		mockSetup   func()
		expectError string
	}{
		{
			name:     "success",
			username: "validuser",
			password: "correctpassword",
			mockSetup: func() {
				mockStorage.On("GetUser", ctx, "validuser").
					Return(validUser, nil).
					Once()
			},
		},
		{
			name:     "user not found",
			username: "unknownuser",
			password: "anypassword",
			mockSetup: func() {
				mockStorage.On("GetUser", ctx, "unknownuser").
					Return(nil, errors.New("not found")).
					Once()
			},
			expectError: "user not found",
		},
		{
			name:     "invalid password",
			username: "validuser",
			password: "wrongpassword",
			mockSetup: func() {
				mockStorage.On("GetUser", ctx, "validuser").
					Return(validUser, nil).
					Once()
			},
			expectError: "invalid password",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockSetup()
			token, err := service.CreateToken(ctx, tc.username, tc.password)

			if tc.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectError)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify the token can be parsed
				username, err := service.VerifyToken(ctx, token)
				assert.NoError(t, err)
				assert.Equal(t, tc.username, username)
			}
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUserService_VerifyToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockStorage := NewMockUserStorage(t)

	mockStorage.On("GetUser", ctx, "testuser").
		Return(&User{Username: "testuser", Password: mustHashPassword("password")}, nil)

	fixedKey := []byte("fixed-signing-key-for-tests")
	service, err := NewUserService(mockStorage, WithTokenSigningKey(fixedKey))
	require.NoError(t, err)

	// Create a valid token
	validToken, err := service.CreateToken(ctx, "testuser", "password")
	require.NoError(t, err)

	// Create an expired token
	expiredToken := createExpiredToken(t, fixedKey, "testuser")

	testCases := []struct {
		name        string
		token       string
		expectError string
		expectUser  string
	}{
		{
			name:       "valid token",
			token:      validToken,
			expectUser: "testuser",
		},
		{
			name:        "expired token",
			token:       expiredToken,
			expectError: "token has invalid claims: token is expired",
		},
		{
			name:        "invalid signature",
			token:       validToken + "tampered",
			expectError: "signature is invalid",
		},
		{
			name:        "malformed token",
			token:       "invalid.token.format",
			expectError: "token is malformed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			username, err := service.VerifyToken(ctx, tc.token)

			if tc.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectError)
				assert.Empty(t, username)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectUser, username)
			}
		})
	}
}

func mustHashPassword(password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}

func createExpiredToken(t *testing.T, key []byte, username string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		UserName: username,
	})
	tokenString, err := token.SignedString(key)
	require.NoError(t, err)
	return tokenString
}
