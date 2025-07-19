// Package secret holds server-side secrets processing logic.
package secret

import (
	"context"
	"errors"
)

var (
	ErrNotFound  = errors.New("secret not found")
	ErrWrongHash = errors.New("wrong hash")
	ErrNoUser    = errors.New("no username supplied")
)

// Secret is an encrypted secret.
type Secret struct {
	Key  string
	Data []byte
	Hash [32]byte
}

// SecretStorage is an interface for storing and retrieving secrets.
type SecretStorage interface {
	Get(ctx context.Context, username, key string) (Secret, error)
	Put(ctx context.Context, username string, secret Secret, hash [32]byte) error // error expected if hash doesn't match already stored hash
	Delete(ctx context.Context, username, key string, hash [32]byte) error        // error expected if hash doesn't match already stored hash
	List(ctx context.Context, username string) ([]Secret, error)                  // no Data expected, only hashes
}

// Service is a secret service.
type Service struct {
	storage SecretStorage
}

// NewService creates a new secret service.
func NewService(storage SecretStorage) *Service {
	return &Service{storage: storage}
}

// GetSecret retrieves a secret by key.
func (s *Service) GetSecret(ctx context.Context, username, key string) (Secret, error) {
	if username == "" {
		return Secret{}, ErrNoUser
	}

	return s.storage.Get(ctx, username, key)
}

// PutSecret stores a secret.
func (s *Service) PutSecret(ctx context.Context, username string, secret Secret, hash [32]byte) error {
	if username == "" {
		return ErrNoUser
	}

	return s.storage.Put(ctx, username, secret, hash)
}

// DeleteSecret deletes a secret by key.
func (s *Service) DeleteSecret(ctx context.Context, username, key string, hash [32]byte) error {
	if username == "" {
		return ErrNoUser
	}

	return s.storage.Delete(ctx, username, key, hash)
}

// ListSecrets lists all secrets.
func (s *Service) ListSecrets(ctx context.Context, username string) ([]Secret, error) {
	if username == "" {
		return nil, ErrNoUser
	}

	return s.storage.List(ctx, username)
}
