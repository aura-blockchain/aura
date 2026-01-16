package validatorsecurity

import (
	"context"
	"net"
	"testing"

	validatorsecuritypb "github.com/aequitas/aura/proto/aura/validatorsecurity/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type vsQueryServer struct {
	validatorsecuritypb.UnimplementedQueryServer
	params *validatorsecuritypb.ValidatorSecurityParams
	err    error
}

func (s *vsQueryServer) Params(ctx context.Context, req *validatorsecuritypb.QueryParamsRequest) (*validatorsecuritypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &validatorsecuritypb.QueryParamsResponse{Params: *s.params}, nil
}

func startVSQueryServer(t *testing.T, srv validatorsecuritypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	validatorsecuritypb.RegisterQueryServer(grpcServer, srv)

	go grpcServer.Serve(lis)

	return lis.Addr().String(), grpcServer.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startVSQueryServer(t, &vsQueryServer{
		params: &validatorsecuritypb.ValidatorSecurityParams{
			RequireSentryNodes: true,
			MinSentryNodes:     2,
		},
	})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := &Client{
		queryClient: validatorsecuritypb.NewQueryClient(conn),
	}

	resp, err := client.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	expected := assert.AnError
	addr, stop := startVSQueryServer(t, &vsQueryServer{err: expected})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := &Client{
		queryClient: validatorsecuritypb.NewQueryClient(conn),
	}

	_, err = client.GetParams(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expected.Error())
}
