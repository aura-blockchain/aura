package prevalidation

import (
	"context"
	"net"
	"testing"

	prevalidationpb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type prevalidationQueryServer struct {
	prevalidationpb.UnimplementedQueryServer
	estimateResp *prevalidationpb.QueryEstimateGasResponse
	validateResp *prevalidationpb.QueryValidateTransactionResponse
	nonce        uint64
	err          error
}

func (s *prevalidationQueryServer) EstimateGas(ctx context.Context, req *prevalidationpb.QueryEstimateGasRequest) (*prevalidationpb.QueryEstimateGasResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.estimateResp, nil
}

func (s *prevalidationQueryServer) ValidateTransaction(ctx context.Context, req *prevalidationpb.QueryValidateTransactionRequest) (*prevalidationpb.QueryValidateTransactionResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.validateResp, nil
}

func (s *prevalidationQueryServer) GetNonce(ctx context.Context, req *prevalidationpb.QueryGetNonceRequest) (*prevalidationpb.QueryGetNonceResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &prevalidationpb.QueryGetNonceResponse{Nonce: s.nonce}, nil
}

func startPrevalidationServer(t *testing.T, srv prevalidationpb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	prevalidationpb.RegisterQueryServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	return lis.Addr().String(), grpcServer.Stop
}

func TestEstimateGasValidation(t *testing.T) {
	c := &Client{queryClient: nil}
	_, _, err := c.EstimateGas(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params cannot be nil")

	_, _, err = c.EstimateGas(context.Background(), &EstimateGasParams{Sender: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender is required")
}

func TestEstimateGasSuccess(t *testing.T) {
	addr, stop := startPrevalidationServer(t, &prevalidationQueryServer{
		estimateResp: &prevalidationpb.QueryEstimateGasResponse{
			GasEstimate: 50000,
			GasLimit:    60000,
		},
	})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: prevalidationpb.NewQueryClient(conn)}

	estimate, limit, err := c.EstimateGas(context.Background(), &EstimateGasParams{Sender: "aura1sender"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(50000), estimate)
	assert.Equal(t, uint64(60000), limit)
}

func TestValidateTransactionValidation(t *testing.T) {
	c := &Client{queryClient: nil}
	_, err := c.ValidateTransaction(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params cannot be nil")

	_, err = c.ValidateTransaction(context.Background(), &ValidateTransactionParams{Sender: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sender is required")
}

func TestValidateTransactionSuccess(t *testing.T) {
	addr, stop := startPrevalidationServer(t, &prevalidationQueryServer{
		validateResp: &prevalidationpb.QueryValidateTransactionResponse{
			Valid:             true,
			GasEstimate:       70000,
			Error:             "",
			SufficientBalance: true,
		},
	})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: prevalidationpb.NewQueryClient(conn)}

	resp, err := c.ValidateTransaction(context.Background(), &ValidateTransactionParams{Sender: "aura1sender"})
	assert.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, uint64(70000), resp.GasEstimate)
}

func TestGetNonceValidation(t *testing.T) {
	c := &Client{queryClient: nil}
	_, err := c.GetNonce(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "address is required")
}

func TestGetNonceSuccess(t *testing.T) {
	addr, stop := startPrevalidationServer(t, &prevalidationQueryServer{nonce: 9})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: prevalidationpb.NewQueryClient(conn)}

	nonce, err := c.GetNonce(context.Background(), "aura1addr")
	assert.NoError(t, err)
	assert.Equal(t, uint64(9), nonce)
}
