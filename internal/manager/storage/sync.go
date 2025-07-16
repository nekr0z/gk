package storage

import (
	"context"
	"errors"

	"github.com/nekr0z/gk/internal/manager/crypt"
)

var ErrConflict = errors.New("conflict")

// Remote is the interface for the remote storage.
type Remote interface {
	List(ctx context.Context) ([]RemoteListedSecret, error)
	Get(ctx context.Context, key string) (crypt.Data, error)
	Put(ctx context.Context, key string, data crypt.Data, hash [32]byte) error
	Delete(ctx context.Context, key string, hash [32]byte) error
}

// RemoteListedSecret is the struct for the listed secret in the remote storage.
type RemoteListedSecret struct {
	Key  string
	Hash [32]byte
}

// ResolverFunc resolves conflicts between local and remote storage.
type ResolverFunc func(ctx context.Context, local, remote crypt.Data) (crypt.Data, error)

func syncAll(ctx context.Context, localStorage Storage, remote Remote, resolver ResolverFunc) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	localList, err := localStorage.List(ctx)
	if err != nil {
		return err
	}

	remoteList, err := remote.List(ctx)
	if err != nil {
		return err
	}

	for _, remoteSecret := range remoteList {
		if local, ok := localList[remoteSecret.Key]; ok {
			if local.Hash == local.LastKnownServerHash && local.Hash == remoteSecret.Hash {
				// nothing to sync
				delete(localList, remoteSecret.Key)
				continue
			}
		}
		err := sync(ctx, localStorage, remote, resolver, remoteSecret.Key)
		if err != nil {
			return err
		}
		delete(localList, remoteSecret.Key)
	}

	for key := range localList {
		err := sync(ctx, localStorage, remote, resolver, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func sync(ctx context.Context, localStorage Storage, remote Remote, resolver ResolverFunc, key string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	localStored, errLocal := localStorage.Get(ctx, key)
	if errLocal != nil && !errors.Is(errLocal, ErrNotFound) {
		return errLocal
	}

	remoteStored, errRemote := remote.Get(ctx, key)
	if errRemote != nil && !errors.Is(errRemote, ErrNotFound) {
		return errRemote
	}

	if errRemote != nil {
		// remote is empty
		if errLocal != nil {
			// local is empty, nothing to do
			return nil
		}

		if localStored.LastKnownServerHash != [32]byte{} {
			// deleted remotely
			return localStorage.Delete(ctx, key)
		}

		return syncToRemote(ctx, localStorage, remote, resolver, key)
	}

	if errLocal != nil {
		// local is empty
		return localStorage.Put(ctx, key, StoredSecret{
			EncryptedPayload:    remoteStored,
			LastKnownServerHash: remoteStored.Hash,
		})
	}

	if localStored.EncryptedPayload.Hash == remoteStored.Hash {
		// payloads are equal
		if localStored.LastKnownServerHash == localStored.EncryptedPayload.Hash {
			// nothing to do
			return nil
		}

		localStored.LastKnownServerHash = remoteStored.Hash

		return localStorage.Put(ctx, key, localStored)
	}

	// now we have different payloads
	if localStored.LastKnownServerHash == remoteStored.Hash {
		// local is newer
		return syncToRemote(ctx, localStorage, remote, resolver, key)
	}

	if localStored.LastKnownServerHash == localStored.EncryptedPayload.Hash {
		// remote is newer
		return localStorage.Put(ctx, key, StoredSecret{
			EncryptedPayload:    remoteStored,
			LastKnownServerHash: remoteStored.Hash,
		})
	}

	// conflict
	if resolver == nil {
		return ErrConflict
	}

	resolved, err := resolver(ctx, localStored.EncryptedPayload, remoteStored)
	if err != nil {
		return err
	}

	if resolved.Hash == remoteStored.Hash {
		// remote won
		return localStorage.Put(ctx, key, StoredSecret{
			EncryptedPayload:    remoteStored,
			LastKnownServerHash: remoteStored.Hash,
		})
	}

	if resolved.Hash == localStored.EncryptedPayload.Hash {
		// local won
		localStored.LastKnownServerHash = remoteStored.Hash
		if err := localStorage.Put(ctx, key, localStored); err != nil {
			return err
		}

		return syncToRemote(ctx, localStorage, remote, resolver, key)
	}

	// resolved to something new entirely
	if err := localStorage.Put(ctx, key, StoredSecret{
		EncryptedPayload:    resolved,
		LastKnownServerHash: remoteStored.Hash,
	}); err != nil {
		return err
	}

	return syncToRemote(ctx, localStorage, remote, resolver, key)
}

func syncToRemote(ctx context.Context, localStorage Storage, remote Remote, resolver ResolverFunc, key string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	localStored, err := localStorage.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(localStored.EncryptedPayload.Data) == 0 && localStored.EncryptedPayload.Hash == [32]byte{} {
		// has been deleted locally
		if err := remote.Delete(ctx, key, localStored.LastKnownServerHash); err != nil {
			if errors.Is(err, ErrConflict) {
				return sync(ctx, localStorage, remote, resolver, key)
			}
			return err
		}

		return localStorage.Delete(ctx, key)
	}

	if err := remote.Put(ctx, key, localStored.EncryptedPayload, localStored.LastKnownServerHash); err != nil {
		if errors.Is(err, ErrConflict) {
			return sync(ctx, localStorage, remote, resolver, key)
		}
		return err
	}

	return localStorage.Put(ctx, key, StoredSecret{
		EncryptedPayload:    localStored.EncryptedPayload,
		LastKnownServerHash: localStored.EncryptedPayload.Hash,
	})
}
