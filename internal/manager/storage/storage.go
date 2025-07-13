// Package storage provides an abstraction over secret storage.
package storage

import (
	"context"
	"fmt"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/secret"
	"github.com/nekr0z/gk/internal/storage"
)

// Repository stores secrets.
type Repository struct {
	storage    storage.Storage
	passPhrase string
}

// New creates a new repository.
func New(storage storage.Storage, passPhrase string) *Repository {
	return &Repository{
		storage:    storage,
		passPhrase: passPhrase,
	}
}

// Create creates a new secret.
func (r *Repository) Create(ctx context.Context, key string, secret secret.Secret) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	encryptedPayload, err := crypt.Encrypt(secret, r.passPhrase)
	if err != nil {
		return err
	}

	return r.storage.Put(ctx, key, storage.Secret{
		EncryptedPayload: encryptedPayload,
	})
}

// Read reads a secret.
func (r *Repository) Read(ctx context.Context, key string) (secret.Secret, error) {
	if ctx.Err() != nil {
		return secret.Secret{}, ctx.Err()
	}

	storedSecret, err := r.storage.Get(ctx, key)
	if err != nil {
		return secret.Secret{}, err
	}

	payload, err := crypt.Decrypt(storedSecret.EncryptedPayload, r.passPhrase)
	if err != nil {
		return secret.Secret{}, fmt.Errorf("failed to decrypt secret: %w; is the passphrase correct?", err)
	}

	return secret.Unmarshal(payload)
}

// Delete deletes a secret.
func (r *Repository) Delete(ctx context.Context, key string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return r.storage.Delete(ctx, key)
}
