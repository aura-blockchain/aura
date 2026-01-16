package cryptography

import (
	"context"
	"net"
	"testing"

	cryptographypb "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type cryptoQueryServer struct {
	cryptographypb.UnimplementedQueryServer
	params *cryptographypb.Params
	err    error
}

func (s *cryptoQueryServer) Params(ctx context.Context, req *cryptographypb.QueryParamsRequest) (*cryptographypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &cryptographypb.QueryParamsResponse{Params: s.params}, nil
}

func startCryptoServer(t *testing.T, srv cryptographypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	cryptographypb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestGetParamsSuccess(t *testing.T) {
	addr, stop := startCryptoServer(t, &cryptoQueryServer{params: &cryptographypb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: cryptographypb.NewQueryClient(conn)}
	resp, err := c.GetParams(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestGetParamsError(t *testing.T) {
	addr, stop := startCryptoServer(t, &cryptoQueryServer{err: assert.AnError})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: cryptographypb.NewQueryClient(conn)}
	_, err = c.GetParams(context.Background())
	assert.Error(t, err)
}
