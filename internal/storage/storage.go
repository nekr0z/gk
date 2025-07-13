// Package storage abstracts the encrypted secrets storage.
package storage

import (
	"context"

	"github.com/nekr0z/gk/internal/manager/crypt"
)

// Storage is a secrets storage.
type Storage interface {
	Get(context.Context, string) (Secret, error)
	Put(context.Context, string, Secret) error
	Delete(context.Context, string) error
}

// Secret is a secret encrypted and stored.
type Secret struct {
	EncryptedPayload    crypt.Data
	LastKnownServerHash [32]byte
}
