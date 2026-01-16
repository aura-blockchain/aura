package economicsecurity

import (
	"context"
	"net"
	"testing"

	economicsecuritypb "github.com/aequitas/aura/proto/aura/economicsecurity/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type esQueryServer struct {
	economicsecuritypb.UnimplementedQueryServer
	params *economicsecuritypb.Params
	err    error
}

func (s *esQueryServer) Params(ctx context.Context, req *economicsecuritypb.QueryParamsRequest) (*economicsecuritypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &economicsecuritypb.QueryParamsResponse{Params: s.params}, nil
}

func startESServer(t *testing.T, srv economicsecuritypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	economicsecuritypb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startESServer(t, &esQueryServer{params: &economicsecuritypb.Params{}})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: economicsecuritypb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startESServer(t, &esQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: economicsecuritypb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
