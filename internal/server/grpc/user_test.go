package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nekr0z/gk/internal/server/user"
	pb "github.com/nekr0z/gk/pkg/pb"
)

type UserServiceServerTestSuite struct {
	suite.Suite
	ctx         context.Context
	mockUser    *MockUserService
	server      *UserServiceServer
	interceptor grpc.UnaryServerInterceptor
}

func (s *UserServiceServerTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.mockUser = NewMockUserService(s.T())
	s.server = NewUserService(s.mockUser)
	s.interceptor = TokenInterceptor(s.mockUser)
}

func TestUserServiceServerTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceServerTestSuite))
}

func (s *UserServiceServerTestSuite) TestSignup_Success() {
	t := s.T()

	s.mockUser.On("Register", mock.Anything, "testuser", "validPassword").Return(nil)

	_, err := s.server.Signup(s.ctx, &pb.SignupRequest{
		Username: "testuser",
		Password: "validPassword",
	})

	require.NoError(t, err)
	s.mockUser.AssertExpectations(t)
}

func (s *UserServiceServerTestSuite) TestSignup_AlreadyExists() {
	t := s.T()

	s.mockUser.On("Register", mock.Anything, "existing", "pass").Return(user.ErrAlreadyExists)

	_, err := s.server.Signup(s.ctx, &pb.SignupRequest{
		Username: "existing",
		Password: "pass",
	})

	require.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Contains(t, err.Error(), "user already exists")
}

func (s *UserServiceServerTestSuite) TestSignup_InternalError() {
	t := s.T()

	expectedErr := errors.New("storage failure")
	s.mockUser.On("Register", mock.Anything, "user", "pass").Return(expectedErr)

	_, err := s.server.Signup(s.ctx, &pb.SignupRequest{
		Username: "user",
		Password: "pass",
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func (s *UserServiceServerTestSuite) TestLogin_Success() {
	t := s.T()

	expectedToken := "valid-token-123"
	s.mockUser.On("CreateToken", mock.Anything, "valid", "password").Return(expectedToken, nil)

	resp, err := s.server.Login(s.ctx, &pb.LoginRequest{
		Username: "valid",
		Password: "password",
	})

	require.NoError(t, err)
	assert.Equal(t, expectedToken, resp.Token)
}

func (s *UserServiceServerTestSuite) TestLogin_InvalidCredentials() {
	t := s.T()

	s.mockUser.On("CreateToken", mock.Anything, "invalid", "creds").Return("", errors.New("invalid credentials"))

	_, err := s.server.Login(s.ctx, &pb.LoginRequest{
		Username: "invalid",
		Password: "creds",
	})

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func (s *UserServiceServerTestSuite) TestTokenInterceptor_UserServiceMethod() {
	t := s.T()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Method in UserService
	info := &grpc.UnaryServerInfo{FullMethod: "/pb.UserService/SomeMethod"}

	resp, err := s.interceptor(s.ctx, nil, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
	s.mockUser.AssertNotCalled(t, "VerifyToken")
}

func (s *UserServiceServerTestSuite) TestTokenInterceptor_MissingMetadata() {
	t := s.T()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pb.OtherService/Method"}

	_, err := s.interceptor(s.ctx, nil, info, handler)

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "metadata is not provided")
}

func (s *UserServiceServerTestSuite) TestTokenInterceptor_MissingToken() {
	t := s.T()

	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(s.ctx, md)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pb.OtherService/Method"}

	_, err := s.interceptor(ctx, nil, info, handler)

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "token is not provided")
}

func (s *UserServiceServerTestSuite) TestTokenInterceptor_InvalidToken() {
	t := s.T()

	md := metadata.New(map[string]string{"authorization": "invalid-token"})
	ctx := metadata.NewIncomingContext(s.ctx, md)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pb.OtherService/Method"}

	s.mockUser.On("VerifyToken", mock.Anything, "invalid-token").Return("", errors.New("invalid token"))

	_, err := s.interceptor(ctx, nil, info, handler)

	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
	assert.Contains(t, err.Error(), "invalid token")
}

func (s *UserServiceServerTestSuite) TestTokenInterceptor_ValidToken() {
	t := s.T()

	md := metadata.New(map[string]string{"authorization": "valid-token"})
	ctx := metadata.NewIncomingContext(s.ctx, md)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify username is set in context
		md, ok := metadata.FromIncomingContext(ctx)
		require.True(t, ok)
		assert.Equal(t, []string{"testuser"}, md.Get("username"))
		return "success", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/pb.OtherService/Method"}

	s.mockUser.On("VerifyToken", mock.Anything, "valid-token").Return("testuser", nil)

	resp, err := s.interceptor(ctx, nil, info, handler)

	require.NoError(t, err)
	assert.Equal(t, "success", resp)
}
