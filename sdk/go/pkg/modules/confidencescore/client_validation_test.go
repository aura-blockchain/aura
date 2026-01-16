package confidencescore

import (
	"context"
	"net"
	"testing"

	confidencescorepb "github.com/aequitas/aura/proto/aura/confidencescore/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type csQueryServer struct {
	confidencescorepb.UnimplementedQueryServer
	params *confidencescorepb.Params
	err    error
}

func (s *csQueryServer) Params(ctx context.Context, req *confidencescorepb.QueryParamsRequest) (*confidencescorepb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &confidencescorepb.QueryParamsResponse{Params: s.params}, nil
}

func startCSServer(t *testing.T, srv confidencescorepb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	confidencescorepb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startCSServer(t, &csQueryServer{params: &confidencescorepb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: confidencescorepb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startCSServer(t, &csQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: confidencescorepb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
