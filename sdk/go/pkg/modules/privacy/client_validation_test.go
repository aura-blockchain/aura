package privacy

import (
	"context"
	"net"
	"testing"

	privacypb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type privacyQueryServer struct {
	privacypb.UnimplementedQueryServer
	params *privacypb.Params
	err    error
}

func (s *privacyQueryServer) Params(ctx context.Context, req *privacypb.QueryParamsRequest) (*privacypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &privacypb.QueryParamsResponse{Params: *s.params}, nil
}

func startPrivacyServer(t *testing.T, srv privacypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	privacypb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startPrivacyServer(t, &privacyQueryServer{params: &privacypb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: privacypb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startPrivacyServer(t, &privacyQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: privacypb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
