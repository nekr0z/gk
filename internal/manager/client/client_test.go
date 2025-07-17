package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/manager/crypt"
	"github.com/nekr0z/gk/internal/manager/storage"
	"github.com/nekr0z/gk/pkg/pb"
)

func makeHashBytes(s string) []byte {
	b := make([]byte, 32)
	a := []byte(s)
	copy(b, a)
	return b
}

// MockSecretServiceClient mocks the gRPC client interface
type MockSecretServiceClient struct {
	ListHashesFunc   func(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pb.ListHashesResponse, error)
	GetSecretFunc    func(context.Context, *pb.GetSecretRequest, ...grpc.CallOption) (*pb.GetSecretResponse, error)
	PutSecretFunc    func(context.Context, *pb.PutSecretRequest, ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteSecretFunc func(context.Context, *pb.DeleteSecretRequest, ...grpc.CallOption) (*emptypb.Empty, error)
}

func (m *MockSecretServiceClient) ListHashes(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListHashesResponse, error) {
	return m.ListHashesFunc(ctx, in, opts...)
}

func (m *MockSecretServiceClient) GetSecret(ctx context.Context, in *pb.GetSecretRequest, opts ...grpc.CallOption) (*pb.GetSecretResponse, error) {
	return m.GetSecretFunc(ctx, in, opts...)
}

func (m *MockSecretServiceClient) PutSecret(ctx context.Context, in *pb.PutSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.PutSecretFunc(ctx, in, opts...)
}

func (m *MockSecretServiceClient) DeleteSecret(ctx context.Context, in *pb.DeleteSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return m.DeleteSecretFunc(ctx, in, opts...)
}

func TestClient_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			ListHashesFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListHashesResponse, error) {
				return &pb.ListHashesResponse{
					Hashes: []*pb.KeyHash{
						{Key: "key1", Hash: makeHashBytes("hash1")},
						{Key: "key2", Hash: makeHashBytes("hash2")},
					},
				}, nil
			},
		}

		c := &Client{s: mockClient}
		secrets, err := c.List(context.Background())
		require.NoError(t, err)
		require.Len(t, secrets, 2)
		assert.Equal(t, "key1", secrets[0].Key)
		assert.Equal(t, [32]byte{'h', 'a', 's', 'h', '1'}, secrets[0].Hash)
		assert.Equal(t, "key2", secrets[1].Key)
		assert.Equal(t, [32]byte{'h', 'a', 's', 'h', '2'}, secrets[1].Hash)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			ListHashesFunc: func(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*pb.ListHashesResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}

		c := &Client{s: mockClient}
		_, err := c.List(context.Background())
		require.Error(t, err)
	})
}

func TestClient_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			GetSecretFunc: func(ctx context.Context, in *pb.GetSecretRequest, opts ...grpc.CallOption) (*pb.GetSecretResponse, error) {
				require.Equal(t, "key1", in.Key)
				return &pb.GetSecretResponse{
					Data: []byte("data1"),
					Hash: makeHashBytes("hash1"),
				}, nil
			},
		}

		c := &Client{s: mockClient}
		data, err := c.Get(context.Background(), "key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("data1"), data.Data)
		assert.Equal(t, [32]byte{'h', 'a', 's', 'h', '1'}, data.Hash)
	})

	t.Run("not found", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			GetSecretFunc: func(ctx context.Context, in *pb.GetSecretRequest, opts ...grpc.CallOption) (*pb.GetSecretResponse, error) {
				return nil, status.Error(codes.NotFound, "not found")
			},
		}

		c := &Client{s: mockClient}
		_, err := c.Get(context.Background(), "key1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound), "expected ErrNotFound")
	})

	t.Run("other error", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			GetSecretFunc: func(ctx context.Context, in *pb.GetSecretRequest, opts ...grpc.CallOption) (*pb.GetSecretResponse, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}

		c := &Client{s: mockClient}
		_, err := c.Get(context.Background(), "key1")
		require.Error(t, err)
		assert.False(t, errors.Is(err, storage.ErrNotFound))
	})
}

func TestClient_Put(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			PutSecretFunc: func(ctx context.Context, in *pb.PutSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				require.Equal(t, "key1", in.Key)
				require.Equal(t, []byte("data1"), in.Data)
				require.Equal(t, makeHashBytes("knownHash"), in.KnownHash)
				return &emptypb.Empty{}, nil
			},
		}

		c := &Client{s: mockClient}
		err := c.Put(context.Background(), "key1", crypt.Data{Data: []byte("data1")}, [32]byte{'k', 'n', 'o', 'w', 'n', 'H', 'a', 's', 'h'})
		require.NoError(t, err)
	})

	t.Run("conflict", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			PutSecretFunc: func(ctx context.Context, in *pb.PutSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, status.Error(codes.FailedPrecondition, "conflict")
			},
		}

		c := &Client{s: mockClient}
		err := c.Put(context.Background(), "key1", crypt.Data{Data: []byte("data1")}, [32]byte{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrConflict), "expected ErrConflict")
	})

	t.Run("other error", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			PutSecretFunc: func(ctx context.Context, in *pb.PutSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}

		c := &Client{s: mockClient}
		err := c.Put(context.Background(), "key1", crypt.Data{Data: []byte("data1")}, [32]byte{})
		require.Error(t, err)
		assert.False(t, errors.Is(err, storage.ErrConflict))
	})
}

func TestClient_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			DeleteSecretFunc: func(ctx context.Context, in *pb.DeleteSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				require.Equal(t, "key1", in.Key)
				require.Equal(t, makeHashBytes("knownHash"), in.KnownHash)
				return &emptypb.Empty{}, nil
			},
		}

		c := &Client{s: mockClient}
		err := c.Delete(context.Background(), "key1", [32]byte{'k', 'n', 'o', 'w', 'n', 'H', 'a', 's', 'h'})
		require.NoError(t, err)
	})

	t.Run("conflict", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			DeleteSecretFunc: func(ctx context.Context, in *pb.DeleteSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, status.Error(codes.FailedPrecondition, "conflict")
			},
		}

		c := &Client{s: mockClient}
		err := c.Delete(context.Background(), "key1", [32]byte{})
		require.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrConflict), "expected ErrConflict")
	})

	t.Run("other error", func(t *testing.T) {
		mockClient := &MockSecretServiceClient{
			DeleteSecretFunc: func(ctx context.Context, in *pb.DeleteSecretRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
				return nil, status.Error(codes.Internal, "internal error")
			},
		}

		c := &Client{s: mockClient}
		err := c.Delete(context.Background(), "key1", [32]byte{})
		require.Error(t, err)
		assert.False(t, errors.Is(err, storage.ErrConflict))
	})
}

func TestToHashArray(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var expected [32]byte
		result := toHashArray(nil)
		assert.Equal(t, expected, result)
	})

	t.Run("valid 32 bytes", func(t *testing.T) {
		b := make([]byte, 32)
		for i := range b {
			b[i] = byte(i)
		}
		var expected [32]byte
		copy(expected[:], b)
		result := toHashArray(b)
		assert.Equal(t, expected, result)
	})

	t.Run("invalid length", func(t *testing.T) {
		b := make([]byte, 31)
		assert.Panics(t, func() {
			toHashArray(b)
		})
	})
}
