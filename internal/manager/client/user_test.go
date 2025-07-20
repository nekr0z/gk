package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/pkg/pb"
)

func TestClient_Signup(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		c := &Client{
			u:        mockClient,
			username: "testuser",
			password: "testpass",
		}

		mockClient.On("Signup", context.Background(), &pb.SignupRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return(&emptypb.Empty{}, nil)

		err := c.Signup(context.Background())
		require.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		c := &Client{
			u:        mockClient,
			username: "testuser",
			password: "testpass",
		}

		mockClient.EXPECT().Signup(context.Background(), &pb.SignupRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return(&emptypb.Empty{}, status.Error(codes.Internal, "internal error"))

		err := c.Signup(context.Background())
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockClient.AssertExpectations(t)
	})
}

func Test_creds_login(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
		}

		mockClient.On("Login", context.Background(), &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return(&pb.LoginResponse{Token: "test-token"}, nil)

		err := cr.login(context.Background(), mockClient)
		require.NoError(t, err)
		assert.Equal(t, "test-token", cr.token)
		mockClient.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
		}

		mockClient.On("Login", context.Background(), &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return((*pb.LoginResponse)(nil), status.Error(codes.Internal, "internal error"))

		err := cr.login(context.Background(), mockClient)
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Empty(t, cr.token)
		mockClient.AssertExpectations(t)
	})
}

func Test_creds_authInterceptor(t *testing.T) {
	t.Parallel()

	t.Run("with existing token", func(t *testing.T) {
		t.Parallel()

		mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			assert.Equal(t, []string{"existing-token"}, md["authorization"])
			return nil
		}

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
			token:    "existing-token",
		}

		interceptor := cr.authInterceptor(mockClient)
		ctx := context.Background()

		err := interceptor(ctx, "method", nil, nil, nil, mockInvoker)
		require.NoError(t, err)
	})

	t.Run("without token, login success", func(t *testing.T) {
		t.Parallel()

		mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			assert.Equal(t, []string{"new-token"}, md["authorization"])
			return nil
		}

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
		}

		mockClient.On("Login", context.Background(), &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return(&pb.LoginResponse{Token: "new-token"}, nil)

		interceptor := cr.authInterceptor(mockClient)
		ctx := context.Background()
		err := interceptor(ctx, "method", nil, nil, nil, mockInvoker)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("without token, login error", func(t *testing.T) {
		t.Parallel()

		mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
		}

		mockClient.On("Login", mock.Anything, &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return((*pb.LoginResponse)(nil), status.Error(codes.Internal, "login failed"))

		interceptor := cr.authInterceptor(mockClient)
		ctx := context.Background()
		err := interceptor(ctx, "method", nil, nil, nil, mockInvoker)
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockClient.AssertExpectations(t)
	})

	t.Run("token refresh on unauthenticated error", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
			token:    "expired-token",
		}

		// First call returns unauthenticated, then we refresh
		mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			if md, ok := metadata.FromOutgoingContext(ctx); ok && md["authorization"][0] == "expired-token" {
				return status.Error(codes.Unauthenticated, "token expired")
			}

			md, ok := metadata.FromOutgoingContext(ctx)
			require.True(t, ok)
			assert.Equal(t, []string{"new-token"}, md["authorization"])
			return nil
		}

		mockClient.On("Login", mock.Anything, &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return(&pb.LoginResponse{Token: "new-token"}, nil)

		interceptor := cr.authInterceptor(mockClient)
		ctx := context.Background()
		err := interceptor(ctx, "method", nil, nil, nil, mockInvoker)
		require.NoError(t, err)

		mockClient.AssertExpectations(t)
	})

	t.Run("token refresh fails after unauthenticated error", func(t *testing.T) {
		t.Parallel()

		mockClient := pb.NewMockUserServiceClient(t)
		cr := &creds{
			username: "testuser",
			password: "testpass",
			token:    "expired-token",
		}

		mockInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return status.Error(codes.Unauthenticated, "token expired")
		}

		mockClient.EXPECT().Login(mock.Anything, &pb.LoginRequest{
			Username: "testuser",
			Password: "testpass",
		}).Return((*pb.LoginResponse)(nil), status.Error(codes.Internal, "login failed"))

		interceptor := cr.authInterceptor(mockClient)
		ctx := context.Background()
		err := interceptor(ctx, "method", nil, nil, nil, mockInvoker)
		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		mockClient.AssertExpectations(t)
	})
}
