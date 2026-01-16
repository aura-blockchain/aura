package dataregistry

import (
	"context"
	"net"
	"testing"

	dataregistrypb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type drQueryServer struct {
	dataregistrypb.UnimplementedQueryServer
	params *dataregistrypb.Params
	err    error
}

func (s *drQueryServer) Params(ctx context.Context, req *dataregistrypb.QueryParamsRequest) (*dataregistrypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &dataregistrypb.QueryParamsResponse{Params: s.params}, nil
}

func startDRServer(t *testing.T, srv dataregistrypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	dataregistrypb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startDRServer(t, &drQueryServer{params: &dataregistrypb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: dataregistrypb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startDRServer(t, &drQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: dataregistrypb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
