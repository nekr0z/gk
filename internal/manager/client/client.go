// Package client is a client to sync with the server.
package client

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/hash"
	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/pkg/pb"
)

var _ storage.Remote = &Client{}

// Client is a client to sync with the server.
type Client struct {
	s pb.SecretServiceClient
	u pb.UserServiceClient

	username string
	password string
}

// Config is the configuration for the client.
type Config struct {
	Address string

	Username string
	Password string

	Insecure bool
}

// New returns a new client.
func New(ctx context.Context, cfg Config) (*Client, error) {
	var opts []grpc.DialOption
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	}

	userConn, err := grpc.NewClient(cfg.Address, opts...)
	if err != nil {
		return nil, err
	}

	userClient := pb.NewUserServiceClient(userConn)
	cred := &creds{
		username: cfg.Username,
		password: cfg.Password,
	}

	opts = append(opts, grpc.WithUnaryInterceptor(cred.authInterceptor(userClient)))

	conn, err := grpc.NewClient(cfg.Address, opts...)
	if err != nil {
		userConn.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		conn.Close()
		userConn.Close()
	}()

	return &Client{
		s: pb.NewSecretServiceClient(conn),
		u: userClient,

		username: cfg.Username,
		password: cfg.Password,
	}, nil
}

// List returns a list of secrets.
func (c *Client) List(ctx context.Context) ([]storage.RemoteListedSecret, error) {
	resp, err := c.s.ListHashes(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	var secrets []storage.RemoteListedSecret
	for _, secret := range resp.GetHashes() {
		secrets = append(secrets, storage.RemoteListedSecret{
			Key:  secret.Key,
			Hash: hash.SliceToArray(secret.Hash),
		})
	}
	return secrets, nil
}

// Get returns a secret.
func (c *Client) Get(ctx context.Context, key string) (crypt.Data, error) {
	resp, err := c.s.GetSecret(ctx, &pb.GetSecretRequest{
		Key: key,
	})

	if status.Code(err) == codes.NotFound {
		return crypt.Data{}, fmt.Errorf("not found: %w - %w", err, storage.ErrNotFound)
	}

	if err != nil {
		return crypt.Data{}, err
	}

	return crypt.Data{
		Data: resp.GetData(),
		Hash: hash.SliceToArray(resp.GetHash()),
	}, nil
}

// Put puts a secret.
func (c *Client) Put(ctx context.Context, key string, data crypt.Data, knownHash [32]byte) error {
	_, err := c.s.PutSecret(ctx, &pb.PutSecretRequest{
		Key:       key,
		Data:      data.Data,
		KnownHash: knownHash[:],
	})

	if status.Code(err) == codes.FailedPrecondition {
		return fmt.Errorf("conflict: %w - %w", err, storage.ErrConflict)
	}

	return err
}

// Delete deletes a secret.
func (c *Client) Delete(ctx context.Context, key string, knownHash [32]byte) error {
	_, err := c.s.DeleteSecret(ctx, &pb.DeleteSecretRequest{
		Key:       key,
		KnownHash: knownHash[:],
	})

	if status.Code(err) == codes.FailedPrecondition {
		return fmt.Errorf("conflict: %w - %w", err, storage.ErrConflict)
	}

	return err
}
