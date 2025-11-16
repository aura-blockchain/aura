package vcregistry

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/aequitas/aura/chain/x/vcregistry/keeper"
	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

type queryServer struct {
	types.UnimplementedQueryServer
	vcregistrypb.UnimplementedQueryServer
	keeper *keeper.Keeper
}

// NewQueryServer creates a new query server
func NewQueryServer(k *keeper.Keeper) vcregistrypb.QueryServer {
	return &queryServer{keeper: k}
}

// GetVC retrieves a specific VC by ID
func (s *queryServer) GetVC(ctx context.Context, req *vcregistrypb.QueryGetVCRequest) (*vcregistrypb.QueryGetVCResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.VcId == "" {
		return nil, fmt.Errorf("vc_id required")
	}

	record, exists := s.keeper.GetVCRecord(req.VcId)
	if !exists {
		return &vcregistrypb.QueryGetVCResponse{
			Vc:     nil,
			Exists: false,
		}, nil
	}

	return &vcregistrypb.QueryGetVCResponse{
		Vc:     types.VCRecordToProto(record),
		Exists: true,
	}, nil
}

// ListUserVCs lists all VCs for a user with optional filters
func (s *queryServer) ListUserVCs(ctx context.Context, req *vcregistrypb.QueryListUserVCsRequest) (*vcregistrypb.QueryListUserVCsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.HolderAddress == "" {
		return nil, fmt.Errorf("holder_address required")
	}

	// Convert proto filters to internal types
	statusFilter := req.StatusFilter
	typeFilter := req.TypeFilter

	// Get VCs from keeper
	vcs := s.keeper.ListUserVCs(req.HolderAddress, statusFilter, typeFilter)

	// Apply pagination
	offset := 0
	limit := 100
	if req.Pagination != nil {
		offset = int(req.Pagination.Offset)
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
	}

	total := len(vcs)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedVCs := vcs[offset:end]

	// Convert to proto
	vcsProto := make([]*vcregistrypb.VCRecord, len(paginatedVCs))
	for i, vc := range paginatedVCs {
		vcsProto[i] = types.VCRecordToProto(vc)
	}

	return &vcregistrypb.QueryListUserVCsResponse{
		Vcs: vcsProto,
		Pagination: &query.PageResponse{
			Total: uint64(total),
		},
	}, nil
}

// CheckVCStatus checks if a VC is valid and returns its status
func (s *queryServer) CheckVCStatus(ctx context.Context, req *vcregistrypb.QueryCheckVCStatusRequest) (*vcregistrypb.QueryCheckVCStatusResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.VcId == "" {
		return nil, fmt.Errorf("vc_id required")
	}

	status, valid, err := s.keeper.CheckVCStatus(req.VcId)
	if err != nil {
		return nil, err
	}

	// Get VC record for expiration time
	record, _ := s.keeper.GetVCRecord(req.VcId)
	expiresAt := record.ExpiresAt

	// Check for revocation record
	var revocationProto *vcregistrypb.RevocationRecord
	if revRecord, ok := s.keeper.GetRevocationRecord(req.VcId); ok {
		revocationProto = types.RevocationRecordToProto(revRecord)
	}

	// Generate Merkle proof (simplified for now)
	merkleProof := s.generateMerkleProof(req.VcId)

	return &vcregistrypb.QueryCheckVCStatusResponse{
		Status:      status,
		Valid:       valid,
		ExpiresAt:   expiresAt,
		Revocation:  revocationProto,
		MerkleProof: merkleProof,
	}, nil
}

// BatchVCStatus checks multiple VCs at once
func (s *queryServer) BatchVCStatus(ctx context.Context, req *vcregistrypb.QueryBatchVCStatusRequest) (*vcregistrypb.QueryBatchVCStatusResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	statuses := make(map[string]*vcregistrypb.VCStatusInfo)

	for _, vcID := range req.VcIds {
		status, valid, err := s.keeper.CheckVCStatus(vcID)
		if err != nil {
			// Skip VCs that don't exist or have errors
			continue
		}

		// Get expiration time
		record, _ := s.keeper.GetVCRecord(vcID)
		expiresAt := record.ExpiresAt

		statuses[vcID] = &vcregistrypb.VCStatusInfo{
			VcId:      vcID,
			Status:    status,
			Valid:     valid,
			ExpiresAt: expiresAt,
		}
	}

	return &vcregistrypb.QueryBatchVCStatusResponse{
		Statuses: statuses,
	}, nil
}

// GetVCPolicy retrieves a VC policy by type name
func (s *queryServer) GetVCPolicy(ctx context.Context, req *vcregistrypb.QueryGetVCPolicyRequest) (*vcregistrypb.QueryGetVCPolicyResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.VcTypeName == "" {
		return nil, fmt.Errorf("vc_type_name required")
	}

	policy, exists := s.keeper.GetVCPolicy(req.VcTypeName)
	if !exists {
		return &vcregistrypb.QueryGetVCPolicyResponse{
			Policy: nil,
			Exists: false,
		}, nil
	}

	return &vcregistrypb.QueryGetVCPolicyResponse{
		Policy: types.VCPolicyToProto(policy),
		Exists: true,
	}, nil
}

// ListVCPolicies lists all VC policies with optional status filter
func (s *queryServer) ListVCPolicies(ctx context.Context, req *vcregistrypb.QueryListVCPoliciesRequest) (*vcregistrypb.QueryListVCPoliciesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	// Convert status filter
	statusFilter := req.StatusFilter

	// Get policies from keeper
	policies := s.keeper.ListVCPolicies(statusFilter)

	// Apply pagination
	offset := 0
	limit := 100
	if req.Pagination != nil {
		offset = int(req.Pagination.Offset)
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
	}

	total := len(policies)
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	paginatedPolicies := policies[offset:end]

	// Convert to proto
	policiesProto := make([]*vcregistrypb.VCPolicy, len(paginatedPolicies))
	for i, policy := range paginatedPolicies {
		policiesProto[i] = types.VCPolicyToProto(policy)
	}

	return &vcregistrypb.QueryListVCPoliciesResponse{
		Policies: policiesProto,
		Pagination: &query.PageResponse{
			Total: uint64(total),
		},
	}, nil
}

// GetRevocationList retrieves the current revocation list with Merkle root
func (s *queryServer) GetRevocationList(ctx context.Context, req *vcregistrypb.QueryGetRevocationListRequest) (*vcregistrypb.QueryGetRevocationListResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	revocationList := s.keeper.GetRevocationList()

	return &vcregistrypb.QueryGetRevocationListResponse{
		RevocationList: types.RevocationListToProto(revocationList),
	}, nil
}

// CheckRevocation checks if a VC is revoked and returns revocation details
func (s *queryServer) CheckRevocation(ctx context.Context, req *vcregistrypb.QueryCheckRevocationRequest) (*vcregistrypb.QueryCheckRevocationResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.VcId == "" {
		return nil, fmt.Errorf("vc_id required")
	}

	revoked := s.keeper.IsRevoked(req.VcId)

	var recordProto *vcregistrypb.RevocationRecord
	var merkleProof []byte

	if revoked {
		if record, ok := s.keeper.GetRevocationRecord(req.VcId); ok {
			recordProto = types.RevocationRecordToProto(record)
			merkleProof = s.generateMerkleProof(req.VcId)
		}
	}

	return &vcregistrypb.QueryCheckRevocationResponse{
		Revoked:     revoked,
		Record:      recordProto,
		MerkleProof: merkleProof,
	}, nil
}

// ResolveDID resolves a DID to its document and associated credentials
func (s *queryServer) ResolveDID(ctx context.Context, req *vcregistrypb.QueryResolveDIDRequest) (*vcregistrypb.QueryResolveDIDResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.Did == "" {
		return nil, fmt.Errorf("did required")
	}

	doc, exists := s.keeper.GetDIDDocument(req.Did)
	if !exists {
		return &vcregistrypb.QueryResolveDIDResponse{
			DidDocument: nil,
			Exists:      false,
			Credentials: []*vcregistrypb.VCRecord{},
		}, nil
	}

	// Get associated active VCs
	credentials := []*vcregistrypb.VCRecord{}
	for _, vcID := range doc.CredentialIds {
		if record, ok := s.keeper.GetVCRecord(vcID); ok {
			if record.Status == types.VCStatusActive {
				credentials = append(credentials, types.VCRecordToProto(record))
			}
		}
	}

	return &vcregistrypb.QueryResolveDIDResponse{
		DidDocument: types.DIDDocumentToProto(doc),
		Exists:      true,
		Credentials: credentials,
	}, nil
}

// GetDIDByAddress retrieves all DIDs for a controller address
func (s *queryServer) GetDIDByAddress(ctx context.Context, req *vcregistrypb.QueryGetDIDByAddressRequest) (*vcregistrypb.QueryGetDIDByAddressResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.Controller == "" {
		return nil, fmt.Errorf("controller required")
	}

	dids := s.keeper.GetDIDsByAddress(req.Controller)
	if dids == nil {
		dids = []string{}
	}

	return &vcregistrypb.QueryGetDIDByAddressResponse{
		Dids: dids,
	}, nil
}

// ValidateMintEligibility checks if a user can mint a specific VC type
func (s *queryServer) ValidateMintEligibility(ctx context.Context, req *vcregistrypb.QueryValidateMintEligibilityRequest) (*vcregistrypb.QueryValidateMintEligibilityResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	if req.HolderAddress == "" {
		return nil, fmt.Errorf("holder_address required")
	}

	// Get VC type name
	vcTypeName := req.VcTypeCustom
	if req.VcType != vcregistrypb.VCType_VC_TYPE_CUSTOM {
		vcTypeName = req.VcType.String()
	}

	// Get policy
	policy, policyExists := s.keeper.GetVCPolicy(vcTypeName)
	if !policyExists {
		return &vcregistrypb.QueryValidateMintEligibilityResponse{
			Eligible:            false,
			MissingRequirements: []string{"Policy not found for VC type"},
			CurrentCs:           0,
			RequiredCs:          0,
			CompletedIrIds:      []string{},
			RequiredIrIds:       []string{},
		}, nil
	}

	// Initialize response
	response := &vcregistrypb.QueryValidateMintEligibilityResponse{
		Eligible:            true,
		MissingRequirements: []string{},
		RequiredCs:          policy.CsThreshold,
		RequiredIrIds:       policy.RequiredIrIds,
		CompletedIrIds:      []string{},
	}

	// Get confidence score keeper
	csKeeper := s.keeper.GetConfidenceScoreKeeper()
	if csKeeper == nil {
		response.Eligible = false
		response.MissingRequirements = append(response.MissingRequirements, "Confidence score system not available")
		return response, nil
	}

	// Check confidence score
	currentCS, _ := csKeeper.GetUserScore(req.HolderAddress)
	response.CurrentCs = currentCS

	if currentCS < policy.CsThreshold {
		response.Eligible = false
		response.MissingRequirements = append(response.MissingRequirements,
			fmt.Sprintf("Insufficient confidence score: %d (required: %d)", currentCS, policy.CsThreshold))
	}

	// Check required IRs
	for _, irID := range policy.RequiredIrIds {
		if csKeeper.HasCompletedIR(req.HolderAddress, irID) {
			response.CompletedIrIds = append(response.CompletedIrIds, irID)
		} else {
			response.Eligible = false
			response.MissingRequirements = append(response.MissingRequirements,
				fmt.Sprintf("Missing required IR: %s", irID))
		}
	}

	// Check arena requirements
	if policy.RequiredArena != "" {
		arenaScore, err := csKeeper.GetArenaScore(req.HolderAddress, policy.RequiredArena)
		if err != nil || arenaScore < policy.RequiredArenaScore {
			response.Eligible = false
			response.MissingRequirements = append(response.MissingRequirements,
				fmt.Sprintf("Insufficient arena score for %s: %d (required: %d)",
					policy.RequiredArena, arenaScore, policy.RequiredArenaScore))
		}
	}

	// Check singleton constraint
	if policy.Singleton {
		existingVCs := s.keeper.ListUserVCs(req.HolderAddress, types.VCStatusActive, req.VcType)
		if len(existingVCs) > 0 {
			response.Eligible = false
			response.MissingRequirements = append(response.MissingRequirements,
				"Singleton VC: user already has an active VC of this type")
		}
	}

	return response, nil
}

// Stats retrieves registry statistics
func (s *queryServer) Stats(ctx context.Context, req *vcregistrypb.QueryStatsRequest) (*vcregistrypb.QueryStatsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	stats := s.keeper.GetStats()

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

// Params retrieves module parameters
func (s *queryServer) Params(ctx context.Context, req *vcregistrypb.QueryParamsRequest) (*vcregistrypb.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request required")
	}

	params := s.keeper.GetParams()

	return &vcregistrypb.QueryParamsResponse{
		Params: types.ParamsToProto(params),
	}, nil
}

// generateMerkleProof generates a Merkle proof for a VC
// This is a simplified implementation - in production, use a proper Merkle tree library
func (s *queryServer) generateMerkleProof(vcID string) []byte {
	h := sha256.New()
	h.Write([]byte(vcID))
	h.Write(s.keeper.GetRevocationList().MerkleRoot)
	return h.Sum(nil)
}
