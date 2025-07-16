// Package storage provides an abstraction over secret storage.
package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/secret"
)

var ErrNotFound = fmt.Errorf("not found")

// Repository stores secrets.
type Repository struct {
	storage    Storage
	remote     Remote
	resolver   ResolverFunc
	passPhrase string
}

// New creates a new repository.
func New(storage Storage, passPhrase string, opts ...Option) (*Repository, error) {
	if storage == nil {
		return nil, errors.New("storage is nil")
	}

	r := &Repository{
		storage:    storage,
		passPhrase: passPhrase,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

// Option is a function that configures a repository.
type Option func(*Repository)

// UseRemote sets the remote storage.
func UseRemote(remote Remote) Option {
	return func(r *Repository) {
		r.remote = remote
	}
}

// UseResolver sets the conflict resolver.
func UseResolver(resolver ResolverFunc) Option {
	return func(r *Repository) {
		r.resolver = resolver
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

	return r.storage.Put(ctx, key, StoredSecret{
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

	if len(storedSecret.EncryptedPayload.Data) == 0 && storedSecret.LastKnownServerHash == [32]byte{} {
		return secret.Secret{}, fmt.Errorf("secret not found")
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

	current, err := r.storage.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			// already deleted
			return nil
		}
		return err
	}

	if current.LastKnownServerHash == [32]byte{} {
		return r.storage.Delete(ctx, key)
	}

	current.EncryptedPayload = crypt.Data{}
	return r.storage.Put(ctx, key, current)
}

// Sync syncs the key with the remote.
func (r *Repository) Sync(ctx context.Context, key string) error {
	if r.remote == nil {
		return fmt.Errorf("remote storage is not set")
	}

	return sync(ctx, r.storage, r.remote, r.resolver, key)
}

// SyncAll syncs all keys with the remote.
func (r *Repository) SyncAll(ctx context.Context) error {
	if r.remote == nil {
		return fmt.Errorf("remote storage is not set")
	}

	return syncAll(ctx, r.storage, r.remote, r.resolver)
}

// Storage is a secrets storage.
type Storage interface {
	Get(context.Context, string) (StoredSecret, error)
	Put(context.Context, string, StoredSecret) error
	Delete(context.Context, string) error
	List(context.Context) (map[string]ListedSecret, error)
}

// StoredSecret is a secret encrypted and stored.
type StoredSecret struct {
	EncryptedPayload    crypt.Data
	LastKnownServerHash [32]byte
}

// ListedSecret is a secret in list.
type ListedSecret struct {
	Hash                [32]byte
	LastKnownServerHash [32]byte
}
