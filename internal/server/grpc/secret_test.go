package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/server/secret"
	pb "github.com/nekr0z/gk/pkg/pb"
)

type SecretServiceServerTestSuite struct {
	suite.Suite
	ctx     context.Context
	mockSec *MockSecretService
	server  *SecretServiceServer
}

func (s *SecretServiceServerTestSuite) SetupTest() {
	s.ctx = authContext()
	s.mockSec = NewMockSecretService(s.T())
	s.server = NewSecretServiceServer(s.mockSec)
}

func authContext() context.Context {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"username": []string{"testuser"},
	})
	return ctx
}

func TestSecretServiceServerTestSuite(t *testing.T) {
	suite.Run(t, new(SecretServiceServerTestSuite))
}

func (s *SecretServiceServerTestSuite) TestGetSecret_Success() {
	t := s.T()

	expectedSecret := secret.Secret{
		Key:  "testkey",
		Data: []byte("secret data"),
		Hash: [32]byte{1, 2, 3, 4},
	}
	s.mockSec.On("GetSecret", s.ctx, "testuser", "testkey").Return(expectedSecret, nil)

	resp, err := s.server.GetSecret(s.ctx, &pb.GetSecretRequest{Key: "testkey"})

	require.NoError(t, err)
	assert.Equal(t, expectedSecret.Data, resp.Data)
	assert.Equal(t, expectedSecret.Hash[:], resp.Hash)
	s.mockSec.AssertExpectations(t)
}

func (s *SecretServiceServerTestSuite) TestGetSecret_NotFound() {
	t := s.T()

	s.mockSec.On("GetSecret", s.ctx, "testuser", "missing").Return(secret.Secret{}, secret.ErrNotFound)

	_, err := s.server.GetSecret(s.ctx, &pb.GetSecretRequest{Key: "missing"})

	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
	assert.Contains(t, err.Error(), "secret not found")
}

func (s *SecretServiceServerTestSuite) TestGetSecret_InternalError() {
	t := s.T()

	expectedErr := errors.New("storage failure")
	s.mockSec.On("GetSecret", s.ctx, "testuser", "error").Return(secret.Secret{}, expectedErr)

	_, err := s.server.GetSecret(s.ctx, &pb.GetSecretRequest{Key: "error"})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func (s *SecretServiceServerTestSuite) TestGetSecret_Unauthenticated() {
	t := s.T()

	// Context without username
	ctx := context.Background()
	_, err := s.server.GetSecret(ctx, &pb.GetSecretRequest{Key: "testkey"})

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "no username in context")
}

func (s *SecretServiceServerTestSuite) TestPutSecret_Success() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	s.mockSec.On("PutSecret", s.ctx, "testuser", mock.Anything, knownHash).Return(nil)

	_, err := s.server.PutSecret(s.ctx, &pb.PutSecretRequest{
		Key:       "testkey",
		Data:      []byte("new data"),
		KnownHash: knownHash[:],
	})

	require.NoError(t, err)
	s.mockSec.AssertExpectations(t)
}

func (s *SecretServiceServerTestSuite) TestPutSecret_WrongHash() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	s.mockSec.On("PutSecret", s.ctx, "testuser", mock.Anything, knownHash).Return(secret.ErrWrongHash)

	_, err := s.server.PutSecret(s.ctx, &pb.PutSecretRequest{
		Key:       "testkey",
		Data:      []byte("new data"),
		KnownHash: knownHash[:],
	})

	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	assert.Contains(t, err.Error(), "wrong hash")
}

func (s *SecretServiceServerTestSuite) TestPutSecret_InternalError() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	expectedErr := errors.New("storage failure")
	s.mockSec.On("PutSecret", s.ctx, "testuser", mock.Anything, knownHash).Return(expectedErr)

	_, err := s.server.PutSecret(s.ctx, &pb.PutSecretRequest{
		Key:       "testkey",
		Data:      []byte("new data"),
		KnownHash: knownHash[:],
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func (s *SecretServiceServerTestSuite) TestDeleteSecret_Success() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	s.mockSec.On("DeleteSecret", s.ctx, "testuser", "testkey", knownHash).Return(nil)

	_, err := s.server.DeleteSecret(s.ctx, &pb.DeleteSecretRequest{
		Key:       "testkey",
		KnownHash: knownHash[:],
	})

	require.NoError(t, err)
	s.mockSec.AssertExpectations(t)
}

func (s *SecretServiceServerTestSuite) TestDeleteSecret_WrongHash() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	s.mockSec.On("DeleteSecret", s.ctx, "testuser", "testkey", knownHash).Return(secret.ErrWrongHash)

	_, err := s.server.DeleteSecret(s.ctx, &pb.DeleteSecretRequest{
		Key:       "testkey",
		KnownHash: knownHash[:],
	})

	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	assert.Contains(t, err.Error(), "wrong hash")
}

func (s *SecretServiceServerTestSuite) TestDeleteSecret_InternalError() {
	t := s.T()

	knownHash := [32]byte{1, 2, 3, 4}
	expectedErr := errors.New("storage failure")
	s.mockSec.On("DeleteSecret", s.ctx, "testuser", "testkey", knownHash).Return(expectedErr)

	_, err := s.server.DeleteSecret(s.ctx, &pb.DeleteSecretRequest{
		Key:       "testkey",
		KnownHash: knownHash[:],
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func (s *SecretServiceServerTestSuite) TestListHashes_Success() {
	t := s.T()

	secrets := []secret.Secret{
		{Key: "key1", Hash: [32]byte{1, 2, 3}},
		{Key: "key2", Hash: [32]byte{4, 5, 6}},
	}
	s.mockSec.On("ListSecrets", s.ctx, "testuser").Return(secrets, nil)

	resp, err := s.server.ListHashes(s.ctx, &emptypb.Empty{})

	require.NoError(t, err)
	require.Len(t, resp.Hashes, 2)
	assert.Equal(t, "key1", resp.Hashes[0].Key)
	assert.Equal(t, secrets[0].Hash[:], resp.Hashes[0].Hash)
	assert.Equal(t, "key2", resp.Hashes[1].Key)
	assert.Equal(t, secrets[1].Hash[:], resp.Hashes[1].Hash)
}

func (s *SecretServiceServerTestSuite) TestListHashes_InternalError() {
	t := s.T()

	expectedErr := errors.New("storage failure")
	s.mockSec.On("ListSecrets", s.ctx, "testuser").Return(nil, expectedErr)

	_, err := s.server.ListHashes(s.ctx, &emptypb.Empty{})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), expectedErr.Error())
}
