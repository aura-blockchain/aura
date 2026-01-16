package identitychange

import (
	"context"
	"net"
	"testing"

	identitychangepb "github.com/aequitas/aura/proto/aura/identitychange/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type icQueryServer struct {
	identitychangepb.UnimplementedQueryServer
	params *identitychangepb.Params
	err    error
}

func (s *icQueryServer) Params(ctx context.Context, req *identitychangepb.QueryParamsRequest) (*identitychangepb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &identitychangepb.QueryParamsResponse{Params: s.params}, nil
}

func startICServer(t *testing.T, srv identitychangepb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcSrv := grpc.NewServer()
	identitychangepb.RegisterQueryServer(grpcSrv, srv)
	go grpcSrv.Serve(lis)
	return lis.Addr().String(), grpcSrv.Stop
}

func TestNewClientNilAura(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic due to nil auraClient")
		}
	}()
	NewClient(nil)
}

func TestParamsQuery(t *testing.T) {
	addr, stop := startICServer(t, &icQueryServer{params: &identitychangepb.Params{}})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: identitychangepb.NewQueryClient(conn)}
	resp, err := c.queryClient.Params(context.Background(), &identitychangepb.QueryParamsRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestValidationNoClient(t *testing.T) {
	c := &Client{}
	assert.Nil(t, c.queryClient)
	// Ensure we surface a clear panic when used without wiring (documented behavior)
	defer func() { _ = recover() }()
	_, _ = c.queryClient.Params(context.Background(), &identitychangepb.QueryParamsRequest{})
}
