package networksecurity

import (
	"context"
	"net"
	"testing"

	networksecuritypb "github.com/aequitas/aura/proto/aura/networksecurity/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type nsQueryServer struct {
	networksecuritypb.UnimplementedQueryServer
	params *networksecuritypb.Params
	err    error
}

func (s *nsQueryServer) Params(ctx context.Context, req *networksecuritypb.QueryParamsRequest) (*networksecuritypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &networksecuritypb.QueryParamsResponse{Params: *s.params}, nil
}

func startNSServer(t *testing.T, srv networksecuritypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	networksecuritypb.RegisterQueryServer(grpcServer, srv)
	go grpcServer.Serve(lis)
	return lis.Addr().String(), grpcServer.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startNSServer(t, &nsQueryServer{params: &networksecuritypb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: networksecuritypb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startNSServer(t, &nsQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: networksecuritypb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
