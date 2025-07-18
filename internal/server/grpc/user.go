// Package grpc implements grpc server.
package grpc

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/server/user"
	"github.com/nekr0z/gk/pkg/pb"
)

// UserServiceServer implements UserServiceServer.
type UserServiceServer struct {
	us UserService

	pb.UnimplementedUserServiceServer
}

// NewUserService returns a new UserService.
func NewUserService(us UserService) *UserServiceServer {
	return &UserServiceServer{
		us: us,
	}
}

// Signup implements UserServiceServer.Signup.
func (s *UserServiceServer) Signup(ctx context.Context, req *pb.SignupRequest) (*emptypb.Empty, error) {
	err := s.us.Register(ctx, req.GetUsername(), req.GetPassword())
	if err == nil {
		return &emptypb.Empty{}, nil
	}

	if errors.Is(err, user.ErrAlreadyExists) {
		return nil, status.Errorf(codes.AlreadyExists, "user already exists")
	}

	return nil, status.Errorf(codes.Internal, err.Error())
}

// Login implements UserServiceServer.Login.
func (s *UserServiceServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.us.CreateToken(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	return &pb.LoginResponse{Token: token}, nil
}

// TokenInterceptor returns a grpc.UnaryServerInterceptor that checks the token.
func TokenInterceptor(us UserService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.Contains(info.FullMethod, "UserService") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
		}

		token := md.Get("authorization")
		if len(token) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "token is not provided")
		}

		username, err := us.VerifyToken(ctx, token[0])
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		md.Set("username", username)
		ctx = metadata.NewIncomingContext(ctx, md)

		return handler(ctx, req)
	}
}

var _ UserService = &user.UserService{}

// UserService is the interface for user.UserService.
type UserService interface {
	Register(ctx context.Context, username, password string) error
	CreateToken(ctx context.Context, username, password string) (string, error)
	VerifyToken(ctx context.Context, token string) (string, error)
}
