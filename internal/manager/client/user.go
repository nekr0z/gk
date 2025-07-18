package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/nekr0z/gk/pkg/pb"
)

// Signup signs up a user.
func (c *Client) Signup(ctx context.Context) error {
	_, err := c.u.Signup(ctx, &pb.SignupRequest{
		Username: c.username,
		Password: c.password,
	})
	return err
}

type creds struct {
	username string
	password string
	token    string
}

func (cr *creds) login(ctx context.Context, c pb.UserServiceClient) error {
	resp, err := c.Login(ctx, &pb.LoginRequest{
		Username: cr.username,
		Password: cr.password,
	})
	if err != nil {
		return err
	}
	cr.token = resp.Token
	return nil
}

func (cr *creds) authInterceptor(c pb.UserServiceClient) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if cr.token == "" {
			if err := cr.login(ctx, c); err != nil {
				return err
			}
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", cr.token))
		err := invoker(ctx, method, req, resp, cc, opts...)
		if err == nil {
			return nil
		}

		if status.Code(err) == codes.Unauthenticated {
			if err := cr.login(ctx, c); err != nil {
				return err
			}
		}

		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", cr.token))
		return invoker(ctx, method, req, resp, cc, opts...)
	}
}
