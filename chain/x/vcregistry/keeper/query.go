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

func (q *QueryServer) syncMetadata(ctx context.Context) {
	if ctx == nil {
		return
	}
	// q.keeper.SyncContextMetadata(ctx) // Commented out - undefined method
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.QrCodeData == "" {
		return nil, types.ErrInvalidQRCodeData
	}

	// Verify the presentation
	result, err := q.keeper.VerifyPresentation(ctx, req.QrCodeData, req.VerifierAddress)
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	vcRecord, exists := q.keeper.GetVCRecord(ctx, req.VcId)
	if !exists {
		return nil, types.ErrVCNotFound
	}

	return &vcregistrypb.QueryGetVCResponse{
		Vc:     types.VCRecordToProto(&vcRecord),
		Exists: true,
	}, nil
}

func (q *QueryServer) ListUserVCs(
	ctx context.Context,
	req *vcregistrypb.QueryListUserVCsRequest,
) (*vcregistrypb.QueryListUserVCsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}

	// Get VCs with filters (convert protobuf enums to types enums)
	vcs := q.keeper.ListUserVCs(
		ctx,
		req.HolderAddress,
		types.VCStatus(req.StatusFilter),
		types.VCType(req.TypeFilter),
	)

	// Apply pagination
	total := uint64(len(vcs))
	offset := uint64(0)
	limit := uint64(100) // Default limit

	if req.Pagination != nil {
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		// Enforce max limit
		if limit > 1000 {
			limit = 1000
		}
	}

	// Slice results based on pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedVCs := vcs[start:end]

	// Convert to protobuf VCRecords
	vcsPtrs := make([]*vcregistrypb.VCRecord, len(paginatedVCs))
	for i := range paginatedVCs {
		vcsPtrs[i] = types.VCRecordToProto(&paginatedVCs[i])
	}

	// Pagination is applied in the slicing logic above (start:end)
	// Return without pagination response type for now as PageResponse needs proto definition
	// In production, would use cosmos.base.query.v1beta1.PageResponse with nextKey
	return &vcregistrypb.QueryListUserVCsResponse{
		Vcs:        vcsPtrs,
		Pagination: nil, // Pagination logic implemented, response type needs proto update
	}, nil
}

func (q *QueryServer) CheckVCStatus(
	ctx context.Context,
	req *vcregistrypb.QueryCheckVCStatusRequest,
) (*vcregistrypb.QueryCheckVCStatusResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Check status
	status, valid, err := q.keeper.CheckVCStatus(ctx, req.VcId)
	if err != nil {
		return nil, err
	}

	// Get VC record for expiration info
	vcRecord, _ := q.keeper.GetVCRecord(ctx, req.VcId)

	// Get revocation record if revoked
	var revocation *vcregistrypb.RevocationRecord
	if revRecord, ok := q.keeper.GetRevocationRecord(ctx, req.VcId); ok {
		// Convert types.RevocationRecord to vcregistrypb.RevocationRecord
		revocation = &vcregistrypb.RevocationRecord{
			VcId:          revRecord.VcId,
			RevokedAt:     revRecord.RevokedAt,
			RevokedHeight: revRecord.RevokedHeight,
			Reason:        vcregistrypb.RevocationReason(revRecord.Reason),
			Revoker:       revRecord.Revoker,
			Evidence:      revRecord.Evidence,
			MerkleProof:   revRecord.MerkleProof,
		}
	}

	// Generate Merkle proof for revocation verification
	var merkleProof []byte
	if revocation != nil {
		// Generate Merkle proof from the revocation list
		merkleProof = q.keeper.GenerateRevocationMerkleProof(ctx, req.VcId)
		// Update revocation with the proof if not already set
		if len(revocation.MerkleProof) == 0 {
			revocation.MerkleProof = merkleProof
		}
	}

	return &vcregistrypb.QueryCheckVCStatusResponse{
		Status:      vcregistrypb.VCStatus(status),
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if len(req.VcIds) == 0 {
		return nil, types.ErrEmptyVCList
	}

	statuses := make(map[string]*vcregistrypb.VCStatusInfo)

	for _, vcID := range req.VcIds {
		status, valid, err := q.keeper.CheckVCStatus(ctx, vcID)
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

		vcRecord, _ := q.keeper.GetVCRecord(ctx, vcID)

		statuses[vcID] = &vcregistrypb.VCStatusInfo{
			VcId:      vcID,
			Status:    vcregistrypb.VCStatus(status),
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	policy, exists := q.keeper.GetVCPolicy(ctx, req.VcTypeName)
	if !exists {
		return nil, types.ErrPolicyNotFound
	}

	return &vcregistrypb.QueryGetVCPolicyResponse{
		Policy: types.VCPolicyToProto(&policy),
		Exists: true,
	}, nil
}

func (q *QueryServer) ListVCPolicies(
	ctx context.Context,
	req *vcregistrypb.QueryListVCPoliciesRequest,
) (*vcregistrypb.QueryListVCPoliciesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	// Get policies with optional status filter
	policies := q.keeper.ListVcPolicies(ctx, types.VCPolicyStatus(req.StatusFilter))

	// Apply pagination
	total := uint64(len(policies))
	offset := uint64(0)
	limit := uint64(50) // Default limit for policies

	if req.Pagination != nil {
		if req.Pagination.Offset > 0 {
			offset = req.Pagination.Offset
		}
		if req.Pagination.Limit > 0 {
			limit = req.Pagination.Limit
		}
		// Enforce max limit
		if limit > 500 {
			limit = 500
		}
	}

	// Slice results based on pagination
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedPolicies := policies[start:end]

	// Convert to protobuf VCPolicies
	policiesPtrs := make([]*vcregistrypb.VCPolicy, len(paginatedPolicies))
	for i := range paginatedPolicies {
		policiesPtrs[i] = types.VCPolicyToProto(&paginatedPolicies[i])
	}

	// Pagination is applied in the slicing logic above (start:end)
	// Return without pagination response type for now as PageResponse needs proto definition
	return &vcregistrypb.QueryListVCPoliciesResponse{
		Policies:   policiesPtrs,
		Pagination: nil, // Pagination logic implemented, response type needs proto update
	}, nil
}

func (q *QueryServer) GetRevocationList(
	ctx context.Context,
	req *vcregistrypb.QueryGetRevocationListRequest,
) (*vcregistrypb.QueryGetRevocationListResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	revocationList := q.keeper.GetRevocationList(ctx)

	// Convert types.RevocationList to vcregistrypb.RevocationList
	pbRevocationList := &vcregistrypb.RevocationList{
		MerkleRoot:        revocationList.MerkleRoot,
		TotalRevocations:  revocationList.TotalRevocations,
		LastUpdatedHeight: revocationList.LastUpdatedHeight,
		LastUpdated:       revocationList.LastUpdated,
	}

	return &vcregistrypb.QueryGetRevocationListResponse{
		RevocationList: pbRevocationList,
	}, nil
}

func (q *QueryServer) CheckRevocation(
	ctx context.Context,
	req *vcregistrypb.QueryCheckRevocationRequest,
) (*vcregistrypb.QueryCheckRevocationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	revoked := q.keeper.IsRevoked(ctx, req.VcId)

	var record *vcregistrypb.RevocationRecord
	var merkleProof []byte

	if revoked {
		if revRecord, ok := q.keeper.GetRevocationRecord(ctx, req.VcId); ok {
			// Generate Merkle proof for this revocation
			merkleProof = q.keeper.GenerateRevocationMerkleProof(ctx, req.VcId)

			// Convert types.RevocationRecord to vcregistrypb.RevocationRecord
			record = &vcregistrypb.RevocationRecord{
				VcId:          revRecord.VcId,
				RevokedAt:     revRecord.RevokedAt,
				RevokedHeight: revRecord.RevokedHeight,
				Reason:        vcregistrypb.RevocationReason(revRecord.Reason),
				Revoker:       revRecord.Revoker,
				Evidence:      revRecord.Evidence,
				MerkleProof:   merkleProof, // Use generated proof
			}
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.Did == "" {
		return nil, types.ErrInvalidDID
	}

	didDoc, exists := q.keeper.GetDIDDocument(ctx, req.Did)
	if !exists {
		return nil, types.ErrDIDNotFound
	}

	// Get associated active credentials
	credentials := []*vcregistrypb.VCRecord{}
	for _, vcID := range didDoc.CredentialIds {
		if vcRecord, ok := q.keeper.GetVCRecord(ctx, vcID); ok {
			if vcRecord.Status == types.VCStatus_VC_STATUS_ACTIVE {
				credentials = append(credentials, types.VCRecordToProto(&vcRecord))
			}
		}
	}

	return &vcregistrypb.QueryResolveDIDResponse{
		DidDocument: types.DIDDocumentToProto(&didDoc),
		Exists:      true,
		Credentials: credentials,
	}, nil
}

func (q *QueryServer) GetDIDByAddress(
	ctx context.Context,
	req *vcregistrypb.QueryGetDIDByAddressRequest,
) (*vcregistrypb.QueryGetDIDByAddressResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.Controller == "" {
		return nil, types.ErrInvalidController
	}

	dids := q.keeper.GetDIDsByAddress(ctx, req.Controller)

	return &vcregistrypb.QueryGetDIDByAddressResponse{
		Dids: dids,
	}, nil
}

func (q *QueryServer) ValidateMintEligibility(
	ctx context.Context,
	req *vcregistrypb.QueryValidateMintEligibilityRequest,
) (*vcregistrypb.QueryValidateMintEligibilityResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	if req.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if req.VcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidVCType
	}

	// Check eligibility (convert protobuf VCType to types.VCType)
	eligible, missingReqs, err := q.keeper.ValidateMintEligibility(ctx, req.HolderAddress, types.VCType(req.VcType))
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

	if policy, ok := q.keeper.GetVCPolicy(ctx, vcTypeName); ok {
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
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	stats := q.keeper.GetStats(ctx)

	// Convert map[types.VCType]uint64 to map[string]uint64 for protobuf response
	vcsByType := make(map[string]uint64)
	for vcType, count := range stats.VCsByType {
		vcsByType[vcType.String()] = count
	}

	return &vcregistrypb.QueryStatsResponse{
		TotalVcsMinted:  stats.TotalVCs,
		TotalActiveVcs:  stats.ActiveVCs,
		TotalRevokedVcs: stats.RevokedVCs,
		TotalExpiredVcs: stats.ExpiredVCs,
		TotalDids:       stats.TotalDIDs,
		TotalPolicies:   stats.TotalPolicies,
		VcsByType:       vcsByType,
	}, nil
}

func (q *QueryServer) Params(
	ctx context.Context,
	req *vcregistrypb.QueryParamsRequest,
) (*vcregistrypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	q.syncMetadata(ctx)

	params := q.keeper.GetParams()

	return &vcregistrypb.QueryParamsResponse{
		Params: types.ParamsToProto(&params),
	}, nil
}

// mustEmbedUnimplementedQueryServer implements the proto interface requirement
//nolint:unused // kept to satisfy generated gRPC interface
func (*QueryServer) mustEmbedUnimplementedQueryServer() {}
