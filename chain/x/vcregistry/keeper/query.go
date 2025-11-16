package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
)

// QueryServer implements the Query service
type QueryServer struct {
	vcregistrypb.UnimplementedQueryServer
	keeper *Keeper
}

// NewQueryServer returns a new QueryServer
func NewQueryServer(keeper *Keeper) *QueryServer {
	return &QueryServer{keeper: keeper}
}

var _ vcregistrypb.QueryServer = &QueryServer{}

// ============================
// PRESENTATION QUERIES
// ============================

// VerifyPresentation handles QueryVerifyPresentation
func (q *QueryServer) VerifyPresentation(
	ctx context.Context,
	req *vcregistrypb.QueryVerifyPresentationRequest,
) (*vcregistrypb.QueryVerifyPresentationResponse, error) {
	if req.QrCodeData == "" {
		return nil, types.ErrInvalidQRCodeData
	}

	// Verify the presentation
	result, err := q.keeper.VerifyPresentation(
		req.QrCodeData,
		req.VerifierAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	return &vcregistrypb.QueryVerifyPresentationResponse{
		Result: result,
	}, nil
}

// Placeholder implementations for other query types
// These would be implemented based on existing vcregistry functionality

func (q *QueryServer) GetVC(
	ctx context.Context,
	req *vcregistrypb.QueryGetVCRequest,
) (*vcregistrypb.QueryGetVCResponse, error) {
	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	vcRecord, exists := q.keeper.GetVCRecord(req.VcId)
	if !exists {
		return &vcregistrypb.QueryGetVCResponse{
			Vc:     nil,
			Exists: false,
		}, nil
	}

	return &vcregistrypb.QueryGetVCResponse{
		Vc:     &vcRecord,
		Exists: true,
	}, nil
}

func (q *QueryServer) ListUserVCs(
	ctx context.Context,
	req *vcregistrypb.QueryListUserVCsRequest,
) (*vcregistrypb.QueryListUserVCsResponse, error) {
	if req.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}

	// Get VCs with filters
	vcs := q.keeper.ListUserVCs(req.HolderAddress, req.StatusFilter, req.TypeFilter)
	// Convert to pointer slice
	vcsPtrs := make([]*vcregistrypb.VCRecord, len(vcs))
	for i := range vcs {
		vcsPtrs[i] = &vcs[i]
	}

	// TODO: Implement pagination using req.Pagination
	// For now, return all results

	return &vcregistrypb.QueryListUserVCsResponse{
		Vcs:        vcsPtrs,
		Pagination: nil, // TODO: implement pagination response
	}, nil
}

func (q *QueryServer) CheckVCStatus(
	ctx context.Context,
	req *vcregistrypb.QueryCheckVCStatusRequest,
) (*vcregistrypb.QueryCheckVCStatusResponse, error) {
	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Check status
	status, valid, err := q.keeper.CheckVCStatus(req.VcId)
	if err != nil {
		return nil, err
	}

	// Get VC record for expiration info
	vcRecord, _ := q.keeper.GetVCRecord(req.VcId)

	// Get revocation record if revoked
	var revocation *vcregistrypb.RevocationRecord
	if revRecord, ok := q.keeper.GetRevocationRecord(req.VcId); ok {
		revocation = &revRecord
	}

	// TODO: Generate Merkle proof for revocation verification
	var merkleProof []byte
	if revocation != nil {
		merkleProof = revocation.MerkleProof
	}

	return &vcregistrypb.QueryCheckVCStatusResponse{
		Status:      status,
		Valid:       valid,
		ExpiresAt:   vcRecord.ExpiresAt,
		Revocation:  revocation,
		MerkleProof: merkleProof,
	}, nil
}

func (q *QueryServer) BatchVCStatus(
	ctx context.Context,
	req *vcregistrypb.QueryBatchVCStatusRequest,
) (*vcregistrypb.QueryBatchVCStatusResponse, error) {
	if len(req.VcIds) == 0 {
		return nil, types.ErrEmptyVCList
	}

	statuses := make(map[string]*vcregistrypb.VCStatusInfo)

	for _, vcID := range req.VcIds {
		status, valid, err := q.keeper.CheckVCStatus(vcID)
		if err != nil {
			// If VC not found, skip it or include error status
			statuses[vcID] = &vcregistrypb.VCStatusInfo{
				VcId:      vcID,
				Status:    vcregistrypb.VCStatus_VC_STATUS_UNSPECIFIED,
				Valid:     false,
				ExpiresAt: nil,
			}
			continue
		}

		vcRecord, _ := q.keeper.GetVCRecord(vcID)

		statuses[vcID] = &vcregistrypb.VCStatusInfo{
			VcId:      vcID,
			Status:    status,
			Valid:     valid,
			ExpiresAt: vcRecord.ExpiresAt,
		}
	}

	return &vcregistrypb.QueryBatchVCStatusResponse{
		Statuses: statuses,
	}, nil
}

func (q *QueryServer) GetVCPolicy(
	ctx context.Context,
	req *vcregistrypb.QueryGetVCPolicyRequest,
) (*vcregistrypb.QueryGetVCPolicyResponse, error) {
	if req.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	policy, exists := q.keeper.GetVCPolicy(req.VcTypeName)
	if !exists {
		return &vcregistrypb.QueryGetVCPolicyResponse{
			Policy: nil,
			Exists: false,
		}, nil
	}

	return &vcregistrypb.QueryGetVCPolicyResponse{
		Policy: &policy,
		Exists: true,
	}, nil
}

func (q *QueryServer) ListVCPolicies(
	ctx context.Context,
	req *vcregistrypb.QueryListVCPoliciesRequest,
) (*vcregistrypb.QueryListVCPoliciesResponse, error) {
	// Get policies with optional status filter
	policies := q.keeper.ListVCPolicies(req.StatusFilter)
	// Convert to pointer slice
	policiesPtrs := make([]*vcregistrypb.VCPolicy, len(policies))
	for i := range policies {
		policiesPtrs[i] = &policies[i]
	}

	// TODO: Implement pagination using req.Pagination
	// For now, return all results

	return &vcregistrypb.QueryListVCPoliciesResponse{
		Policies:   policiesPtrs,
		Pagination: nil, // TODO: implement pagination response
	}, nil
}

func (q *QueryServer) GetRevocationList(
	ctx context.Context,
	req *vcregistrypb.QueryGetRevocationListRequest,
) (*vcregistrypb.QueryGetRevocationListResponse, error) {
	revocationList := q.keeper.GetRevocationList()

	return &vcregistrypb.QueryGetRevocationListResponse{
		RevocationList: revocationList,
	}, nil
}

func (q *QueryServer) CheckRevocation(
	ctx context.Context,
	req *vcregistrypb.QueryCheckRevocationRequest,
) (*vcregistrypb.QueryCheckRevocationResponse, error) {
	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	revoked := q.keeper.IsRevoked(req.VcId)

	var record *vcregistrypb.RevocationRecord
	var merkleProof []byte

	if revoked {
		if revRecord, ok := q.keeper.GetRevocationRecord(req.VcId); ok {
			record = &revRecord
			merkleProof = revRecord.MerkleProof
		}
	}

	return &vcregistrypb.QueryCheckRevocationResponse{
		Revoked:     revoked,
		Record:      record,
		MerkleProof: merkleProof,
	}, nil
}

func (q *QueryServer) ResolveDID(
	ctx context.Context,
	req *vcregistrypb.QueryResolveDIDRequest,
) (*vcregistrypb.QueryResolveDIDResponse, error) {
	if req.Did == "" {
		return nil, types.ErrInvalidDID
	}

	didDoc, exists := q.keeper.GetDIDDocument(req.Did)
	if !exists {
		return &vcregistrypb.QueryResolveDIDResponse{
			DidDocument: nil,
			Exists:      false,
			Credentials: nil,
		}, nil
	}

	// Get associated active credentials
	credentials := []*vcregistrypb.VCRecord{}
	for _, vcID := range didDoc.CredentialIds {
		if vcRecord, ok := q.keeper.GetVCRecord(vcID); ok {
			if vcRecord.Status == vcregistrypb.VCStatus_VC_STATUS_ACTIVE {
				credentials = append(credentials, &vcRecord)
			}
		}
	}

	return &vcregistrypb.QueryResolveDIDResponse{
		DidDocument: &didDoc,
		Exists:      true,
		Credentials: credentials,
	}, nil
}

func (q *QueryServer) GetDIDByAddress(
	ctx context.Context,
	req *vcregistrypb.QueryGetDIDByAddressRequest,
) (*vcregistrypb.QueryGetDIDByAddressResponse, error) {
	if req.Controller == "" {
		return nil, types.ErrInvalidController
	}

	dids := q.keeper.GetDIDsByAddress(req.Controller)

	return &vcregistrypb.QueryGetDIDByAddressResponse{
		Dids: dids,
	}, nil
}

func (q *QueryServer) ValidateMintEligibility(
	ctx context.Context,
	req *vcregistrypb.QueryValidateMintEligibilityRequest,
) (*vcregistrypb.QueryValidateMintEligibilityResponse, error) {
	if req.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if req.VcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidVCType
	}

	// Check eligibility
	eligible, missingReqs, err := q.keeper.ValidateMintEligibility(req.HolderAddress, req.VcType)
	if err != nil {
		return nil, err
	}

	// Get current CS
	var currentCS uint64
	if q.keeper.csKeeper != nil {
		currentCS, _ = q.keeper.csKeeper.GetUserScore(req.HolderAddress)
	}

	// Get policy to return required CS and IRs
	vcTypeName := req.VcTypeCustom
	if req.VcType != vcregistrypb.VCType_VC_TYPE_CUSTOM {
		vcTypeName = fmt.Sprintf("%d", req.VcType)
	}

	var requiredCS uint64
	var requiredIRs []string
	var completedIRs []string

	if policy, ok := q.keeper.GetVCPolicy(vcTypeName); ok {
		requiredCS = policy.CsThreshold
		requiredIRs = policy.RequiredIrIds

		// Check which IRs are completed
		if q.keeper.csKeeper != nil {
			for _, irID := range requiredIRs {
				if q.keeper.csKeeper.HasCompletedIR(req.HolderAddress, irID) {
					completedIRs = append(completedIRs, irID)
				}
			}
		}
	}

	return &vcregistrypb.QueryValidateMintEligibilityResponse{
		Eligible:            eligible,
		MissingRequirements: missingReqs,
		CurrentCs:           currentCS,
		RequiredCs:          requiredCS,
		CompletedIrIds:      completedIRs,
		RequiredIrIds:       requiredIRs,
	}, nil
}

func (q *QueryServer) Stats(
	ctx context.Context,
	req *vcregistrypb.QueryStatsRequest,
) (*vcregistrypb.QueryStatsResponse, error) {
	stats := q.keeper.GetStats()

	return &vcregistrypb.QueryStatsResponse{
		TotalVcsMinted:  stats.TotalVCs,
		TotalActiveVcs:  stats.ActiveVCs,
		TotalRevokedVcs: stats.RevokedVCs,
		TotalExpiredVcs: stats.ExpiredVCs,
		TotalDids:       stats.TotalDIDs,
		TotalPolicies:   stats.TotalPolicies,
		VcsByType:       stats.VCsByType,
	}, nil
}

func (q *QueryServer) Params(
	ctx context.Context,
	req *vcregistrypb.QueryParamsRequest,
) (*vcregistrypb.QueryParamsResponse, error) {
	params := q.keeper.GetParams()

	return &vcregistrypb.QueryParamsResponse{
		Params: types.ParamsToProto(params),
	}, nil
}

// mustEmbedUnimplementedQueryServer implements the proto interface requirement
func (*QueryServer) mustEmbedUnimplementedQueryServer() {}
