// Package user handles user authentication and authorization.
package user

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	TokenLifetime = time.Hour

	ErrAlreadyExists = errors.New("user already exists")
)

const (
	tokenDefaultLen = 32
)

// User is a struct that represents a user.
type User struct {
	Username string
	Password []byte
}

// UserStorage is an interface that represents a storage for users.
type UserStorage interface {
	AddUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, username string) (*User, error)
}

// UserService represents a service for users.
type UserService struct {
	storage         UserStorage
	tokenSigningKey []byte
}

// NewUserService creates a new user service.
func NewUserService(storage UserStorage, opts ...Option) (*UserService, error) {
	s := &UserService{storage: storage}

	for _, opt := range opts {
		opt(s)
	}

	if s.tokenSigningKey == nil {
		s.tokenSigningKey = make([]byte, tokenDefaultLen)
		if _, err := rand.Read(s.tokenSigningKey); err != nil {
			return s, err
		}
	}

	return s, nil
}

// Option is a function that configures a user service.
type Option func(*UserService)

// WithTokenSigningKey sets the token signing key.
func WithTokenSigningKey(key []byte) Option {
	return func(s *UserService) {
		s.tokenSigningKey = key
	}
}

// Register registers a new user.
func (s *UserService) Register(ctx context.Context, username, password string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &User{
		Username: username,
		Password: h,
	}
	return s.storage.AddUser(ctx, user)
}

// CreateToken authenticates the user and creates a new token.
func (s *UserService) CreateToken(ctx context.Context, username, password string) (string, error) {
	user, err := s.storage.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenLifetime)),
		},
		UserName: username,
	})

	tokenString, err := token.SignedString(s.tokenSigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken verifies the token and returns the username.
func (s *UserService) VerifyToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.tokenSigningKey, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", errors.New("no valid claims")
	}

	return claims.UserName, nil
}

// Claims is a struct that will be encoded to a JWT.
type Claims struct {
	UserName string `json:"username"`
	jwt.RegisteredClaims
}
