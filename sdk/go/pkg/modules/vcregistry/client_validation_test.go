package vcregistry

import (
	"context"
	"net"
	"testing"

	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// Minimal query server implementing only methods exercised in tests.
type vcQueryServer struct {
	vcregistrypb.UnimplementedQueryServer
	vcRecord  *vcregistrypb.VCRecord
	vcRecords []*vcregistrypb.VCRecord
	status    *vcregistrypb.QueryCheckVCStatusResponse
	params    *vcregistrypb.Params
	err       error
}

func (s *vcQueryServer) GetVC(ctx context.Context, req *vcregistrypb.QueryGetVCRequest) (*vcregistrypb.QueryGetVCResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &vcregistrypb.QueryGetVCResponse{
		Vc:     s.vcRecord,
		Exists: s.vcRecord != nil,
	}, nil
}

func (s *vcQueryServer) BatchVCStatus(ctx context.Context, req *vcregistrypb.QueryBatchVCStatusRequest) (*vcregistrypb.QueryBatchVCStatusResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	result := map[string]*vcregistrypb.VCStatusInfo{}
	for _, id := range req.VcIds {
		result[id] = &vcregistrypb.VCStatusInfo{Status: vcregistrypb.VCStatus_VC_STATUS_ACTIVE}
	}
	return &vcregistrypb.QueryBatchVCStatusResponse{Statuses: result}, nil
}

func (s *vcQueryServer) CheckVCStatus(ctx context.Context, req *vcregistrypb.QueryCheckVCStatusRequest) (*vcregistrypb.QueryCheckVCStatusResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.status != nil {
		return s.status, nil
	}
	return &vcregistrypb.QueryCheckVCStatusResponse{
		Status: vcregistrypb.VCStatus_VC_STATUS_ACTIVE,
		Valid:  true,
	}, nil
}

func (s *vcQueryServer) ListUserVCs(ctx context.Context, req *vcregistrypb.QueryListUserVCsRequest) (*vcregistrypb.QueryListUserVCsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &vcregistrypb.QueryListUserVCsResponse{Vcs: s.vcRecords}, nil
}

func (s *vcQueryServer) Params(ctx context.Context, req *vcregistrypb.QueryParamsRequest) (*vcregistrypb.QueryParamsResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &vcregistrypb.QueryParamsResponse{Params: s.params}, nil
}

func startVCServer(t *testing.T, srv vcregistrypb.QueryServer) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	vcregistrypb.RegisterQueryServer(grpcServer, srv)
	go grpcServer.Serve(lis)

	return lis.Addr().String(), grpcServer.Stop
}

func TestMintVCValidation(t *testing.T) {
	c := &Client{auraClient: nil}
	_, err := c.MintVC(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params cannot be nil")

	_, err = c.MintVC(context.Background(), &MintVCParams{HolderAddress: "", HolderDID: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "holder address is required")
}

func TestRevokeVCValidation(t *testing.T) {
	c := &Client{auraClient: nil}
	_, err := c.RevokeVC(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params cannot be nil")

	_, err = c.RevokeVC(context.Background(), &RevokeVCParams{HolderAddress: "", VCID: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "holder address is required")
}

func TestGetVCAndStatusQueries(t *testing.T) {
	addr, stop := startVCServer(t, &vcQueryServer{
		vcRecord: &vcregistrypb.VCRecord{VcId: "vc:aura:123"},
		params:   &vcregistrypb.Params{},
	})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: vcregistrypb.NewQueryClient(conn)}

	vc, err := c.GetVC(context.Background(), "vc:aura:123")
	assert.NoError(t, err)
	assert.Equal(t, "vc:aura:123", vc.VcId)

	statusOK, resp, err := c.VerifyVC(context.Background(), "vc:aura:123")
	assert.NoError(t, err)
	assert.True(t, statusOK)
	assert.Equal(t, vcregistrypb.VCStatus_VC_STATUS_ACTIVE, resp.Status)
}

func TestBatchVCStatus(t *testing.T) {
	addr, stop := startVCServer(t, &vcQueryServer{})
	defer stop()
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: vcregistrypb.NewQueryClient(conn)}

	statuses, err := c.BatchVCStatus(context.Background(), []string{"vc1", "vc2"})
	assert.NoError(t, err)
	assert.Len(t, statuses, 2)
}

func TestListVCsValidation(t *testing.T) {
	c := &Client{}
	_, err := c.ListVCs(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "holder address is required")
}

func TestGetVCNotFoundError(t *testing.T) {
	addr, stop := startVCServer(t, &vcQueryServer{vcRecord: nil})
	defer stop()

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	c := &Client{queryClient: vcregistrypb.NewQueryClient(conn)}

	_, err = c.GetVC(context.Background(), "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
