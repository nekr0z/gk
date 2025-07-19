package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/hash"
	"github.com/nekr0z/gk/internal/server/secret"
	"github.com/nekr0z/gk/pkg/pb"
)

// SecretServiceServer is the server API for secrets service.
type SecretServiceServer struct {
	s SecretService

	pb.UnimplementedSecretServiceServer
}

// NewSecretServiceServer creates a new secret service server.
func NewSecretServiceServer(s SecretService) *SecretServiceServer {
	return &SecretServiceServer{s: s}
}

// GetSecret returns a secret.
func (s *SecretServiceServer) GetSecret(ctx context.Context, req *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "no username in context")
	}

	key := req.GetKey()

	sec, err := s.s.GetSecret(ctx, username, key)
	if err == nil {
		return &pb.GetSecretResponse{
			Data: sec.Data,
			Hash: sec.Hash[:],
		}, nil
	}

	if errors.Is(err, secret.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	return nil, status.Errorf(codes.Internal, "internal error: %v", err)
}

// PutSecret stores a secret.
func (s *SecretServiceServer) PutSecret(ctx context.Context, req *pb.PutSecretRequest) (*emptypb.Empty, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "no username in context")
	}

	key := req.GetKey()
	data := req.GetData()

	err = s.s.PutSecret(ctx, username, secret.Secret{
		Key:  key,
		Data: data,
	}, hash.SliceToArray(req.GetKnownHash()))
	if err == nil {
		return &emptypb.Empty{}, nil
	}

	if errors.Is(err, secret.ErrWrongHash) {
		return nil, status.Error(codes.FailedPrecondition, "wrong hash")
	}

	return nil, status.Errorf(codes.Internal, "internal error: %v", err)
}

// DeleteSecret deletes a secret.
func (s *SecretServiceServer) DeleteSecret(ctx context.Context, req *pb.DeleteSecretRequest) (*emptypb.Empty, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "no username in context")
	}

	key := req.GetKey()

	err = s.s.DeleteSecret(ctx, username, key, hash.SliceToArray(req.GetKnownHash()))
	if err == nil {
		return &emptypb.Empty{}, nil
	}

	if errors.Is(err, secret.ErrWrongHash) {
		return nil, status.Error(codes.FailedPrecondition, "wrong hash")
	}

	return nil, status.Errorf(codes.Internal, "internal error: %v", err)
}

// ListHashes lists all known hashes.
func (s *SecretServiceServer) ListHashes(ctx context.Context, _ *emptypb.Empty) (*pb.ListHashesResponse, error) {
	username, err := usernameFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "no username in context")
	}

	hashes, err := s.s.ListSecrets(ctx, username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error: %v", err)
	}

	resp := &pb.ListHashesResponse{}
	for _, h := range hashes {
		resp.Hashes = append(resp.Hashes, &pb.KeyHash{
			Key:  h.Key,
			Hash: h.Hash[:],
		})
	}

	return resp, nil
}

// SecretService is the interface for secret.Service.
type SecretService interface {
	GetSecret(context.Context, string, string) (secret.Secret, error)
	PutSecret(context.Context, string, secret.Secret, [32]byte) error
	DeleteSecret(context.Context, string, string, [32]byte) error
	ListSecrets(context.Context, string) ([]secret.Secret, error)
}

var _ SecretService = (*secret.Service)(nil)

func usernameFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}

	u := md.Get("username")

	if len(u) == 0 {
		return "", errors.New("no username in context")
	}
	return u[0], nil
}
