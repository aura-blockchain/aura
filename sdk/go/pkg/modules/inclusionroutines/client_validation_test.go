package inclusionroutines

import (
	"context"
	"net"
	"testing"

	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type irQueryServer struct {
	inclusionroutinespb.UnimplementedQueryServer
	params *inclusionroutinespb.Params
	err    error
}

func (s *irQueryServer) Params(ctx context.Context, req *inclusionroutinespb.QueryParamsRequest) (*inclusionroutinespb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &inclusionroutinespb.QueryParamsResponse{Params: s.params}, nil
}

func startIRServer(t *testing.T, srv inclusionroutinespb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	inclusionroutinespb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startIRServer(t, &irQueryServer{params: &inclusionroutinespb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: inclusionroutinespb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startIRServer(t, &irQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: inclusionroutinespb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
