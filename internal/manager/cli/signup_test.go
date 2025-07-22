package cli_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nekr0z/gk/internal/manager/cli"
	"github.com/nekr0z/gk/pkg/pb"
)

const (
	username = "testuser"
	password = "testpass"
)

func TestSignup(t *testing.T) {
	dbFilename := ":memory:?cache=shared"

	lis, err := net.Listen("tcp", ":")
	require.NoError(t, err)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &mockUserServer{})

	go func() {
		err := s.Serve(lis)
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		s.Stop()
	})

	cmd := cli.RootCmd()
	os.Setenv("LANGUAGE", "en")

	b := &bytes.Buffer{}
	cmd.SetOut(b)

	cmd.SetArgs([]string{"signup", "-d", dbFilename, "-p", passPhrase, "-s", lis.Addr().String(), "-i", "-u", username, "-w", password})
	cmd.Execute()

	out, err := io.ReadAll(b)
	assert.NoError(t, err)

	assert.Contains(t, string(out), username)
	assert.Contains(t, string(out), "successful")
}

type mockUserServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *mockUserServer) Signup(ctx context.Context, req *pb.SignupRequest) (*emptypb.Empty, error) {
	if req.Username != username {
		return nil, status.Error(codes.FailedPrecondition, "username already taken")
	}

	if req.Password != password {
		return nil, status.Error(codes.FailedPrecondition, "unexpected password")
	}

	return &emptypb.Empty{}, nil
}
