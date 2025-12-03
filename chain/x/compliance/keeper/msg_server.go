package keeper

import (
	context "context"
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

type msgServer struct {
	types.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServer returns a compliance MsgServer implementation backed by the keeper.
func NewMsgServer(k *Keeper) types.MsgServer {
	return &msgServer{Keeper: k}
}

// SubmitKYC stores a KYC record using commitment-based storage (GDPR-compliant).
// The PII data (verification_id, documents, jurisdiction, risk_score) must be
// stored off-chain by the KYC provider. Only the cryptographic commitment (SHA-256 hash)
// of the PII is stored on-chain, allowing verification without storing sensitive data.
//
// Security considerations:
//   - Provider must be authorized (checked against params.ApprovedKycProviders)
//   - Provider must be the transaction signer (authentication)
//   - PII commitment must be exactly 32 bytes (SHA-256 hash)
//   - Off-chain PII data should be encrypted and stored by the provider
//   - Data erasure requests can be fulfilled off-chain while preserving on-chain audit trail
//
// GDPR compliance:
//   - Article 17 "Right to Erasure": PII stored off-chain can be deleted on request
//   - Article 5 "Data Minimization": Only essential verification status stored on-chain
//   - Immutable audit trail maintained via commitment hashes
func (s *msgServer) SubmitKYC(goCtx context.Context, req *types.MsgSubmitKYC) (*types.MsgSubmitKYCResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	if req.Provider == "" {
		return nil, status.Error(codes.InvalidArgument, "provider is required")
	}
	if len(req.PiiCommitment) != 32 {
		return nil, status.Error(codes.InvalidArgument, "pii_commitment must be 32 bytes (SHA-256 hash)")
	}

	// Verify signer
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	providerAddr, err := sdk.AccAddressFromBech32(req.Provider)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid provider address")
	}

	if !providerAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "provider must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if provider is authorized
	params := s.Keeper.GetParams(ctx)
	isAuthorized := false
	for _, authorizedProvider := range params.ApprovedKycProviders {
		if authorizedProvider == req.Provider {
			isAuthorized = true
			break
		}
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, "provider not authorized to submit KYC records")
	}

	now := ctx.BlockTime()
	expiresAt := timestamppb.New(now.Add(time.Duration(params.KycExpiryDays) * 24 * time.Hour))
	record := &types.KYCRecord{
		Address:              req.Address,
		KycLevel:             req.KycLevel,
		Provider:             req.Provider,
		VerifiedAt:           timestamppb.New(now),
		ExpiresAt:            expiresAt,
		PiiCommitment:        req.PiiCommitment,
		EnhancedDueDiligence: req.KycLevel == types.KYCLevel_KYC_LEVEL_ADVANCED,
	}
	if err := s.Keeper.SetKYCRecord(ctx, record); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for KYC submission
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"kyc_submitted",
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("provider", req.Provider),
			sdk.NewAttribute("kyc_level", req.KycLevel.String()),
			sdk.NewAttribute("pii_commitment", fmt.Sprintf("%x", req.PiiCommitment)),
		),
	)

	return &types.MsgSubmitKYCResponse{Success: true, Message: "kyc record stored with PII commitment"}, nil
}

func (s *msgServer) ReportSuspiciousActivity(goCtx context.Context, req *types.MsgReportSuspiciousActivity) (*types.MsgReportSuspiciousActivityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.TransactionHash == "" {
		return nil, status.Error(codes.InvalidArgument, "address and transaction hash required")
	}
	if req.Reporter == "" {
		return nil, status.Error(codes.InvalidArgument, "reporter is required")
	}

	// Verify signer
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	reporterAddr, err := sdk.AccAddressFromBech32(req.Reporter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reporter address")
	}

	if !reporterAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "reporter must be transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()
	id := fmt.Sprintf("sar-%s-%d", req.TransactionHash, now.UnixNano())
	activity := &types.SuspiciousActivity{
		Id:              id,
		Address:         req.Address,
		TransactionHash: req.TransactionHash,
		ActivityType:    req.ActivityType,
		Description:     req.Description,
		DetectedAt:      timestamppb.New(now),
		ReportedAt:      timestamppb.New(now),
		Indicators:      req.Indicators,
		FiledSar:        false,
	}
	if err := s.Keeper.SetSuspiciousActivity(ctx, activity); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for SAR filing
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"suspicious_activity_reported",
			sdk.NewAttribute("activity_id", id),
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("reporter", req.Reporter),
			sdk.NewAttribute("activity_type", req.ActivityType),
		),
	)

	return &types.MsgReportSuspiciousActivityResponse{ActivityId: id}, nil
}

func (s *msgServer) ScreenSanctions(goCtx context.Context, req *types.MsgScreenSanctions) (*types.MsgScreenSanctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	// Verify signer - the address being screened must be the signer (user-initiated screening)
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	requestAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	if !requestAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "address must match transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var result *types.SanctionsScreeningResult
	if !req.ForceRefresh {
		result, err = s.Keeper.GetSanctionsResult(ctx, req.Address)
		if err != nil {
			result = nil
		}
	}
	if result == nil {
		result, err = s.performSanctionsScreen(ctx, req.Address)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := s.Keeper.SetSanctionsResult(ctx, result); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Emit event for sanctions screening
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"sanctions_screened",
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("status", result.Status.String()),
			sdk.NewAttribute("requires_review", fmt.Sprintf("%t", result.RequiresManualReview)),
		),
	)

	return &types.MsgScreenSanctionsResponse{Status: result.Status, RequiresReview: result.RequiresManualReview}, nil
}

func (s *msgServer) performSanctionsScreen(ctx sdk.Context, address string) (*types.SanctionsScreeningResult, error) {
	now := ctx.BlockTime()
	// prefer registered providers if any
	if len(s.Keeper.sanctionsProviders) > 0 {
		// iterate in deterministic order
		names := make([]string, 0, len(s.Keeper.sanctionsProviders))
		for name := range s.Keeper.sanctionsProviders {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			provider := s.Keeper.sanctionsProviders[name]
			res, err := provider.ScreenAddress(address)
			if err == nil && res != nil {
				if res.ScreenedAt == nil {
					res.ScreenedAt = timestamppb.New(now)
				}
				res.ScreeningProvider = name
				return res, nil
			}
		}
	}
	return &types.SanctionsScreeningResult{
		Address:              address,
		Status:               types.SanctionsStatus_SANCTIONS_CLEAR,
		Matches:              []*types.SanctionsMatch{},
		ScreenedAt:           timestamppb.New(now),
		ScreeningProvider:    "internal",
		RequiresManualReview: false,
	}, nil
}

func (s *msgServer) RecordGDPRConsent(goCtx context.Context, req *types.MsgRecordGDPRConsent) (*types.MsgRecordGDPRConsentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.ConsentType == "" {
		return nil, status.Error(codes.InvalidArgument, "address and consent type required")
	}

	// Verify signer - the address giving/withdrawing consent must be the signer
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	requestAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	if !requestAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "address must match transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()
	consent := &types.GDPRConsent{
		Address:        req.Address,
		ConsentType:    req.ConsentType,
		Consented:      req.Consented,
		ConsentVersion: req.ConsentVersion,
		ConsentGivenAt: timestamppb.New(now),
	}
	if !req.Consented {
		consent.ConsentWithdrawnAt = timestamppb.New(now)
	}
	if err := s.Keeper.SetGDPRConsent(ctx, consent); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for GDPR consent
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_consent_recorded",
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("consent_type", req.ConsentType),
			sdk.NewAttribute("consented", fmt.Sprintf("%t", req.Consented)),
		),
	)

	return &types.MsgRecordGDPRConsentResponse{Success: true}, nil
}

func (s *msgServer) RequestGDPRData(goCtx context.Context, req *types.MsgRequestGDPRData) (*types.MsgRequestGDPRDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.RequestType == "" {
		return nil, status.Error(codes.InvalidArgument, "address and request type required")
	}

	// Verify signer - the address requesting data must be the signer
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	requestAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	if !requestAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "address must match transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()
	id := fmt.Sprintf("gdpr-%d-%d", ctx.BlockHeight(), now.UnixNano())
	request := &types.GDPRDataRequest{
		Id:          id,
		Address:     req.Address,
		RequestType: req.RequestType,
		RequestedAt: timestamppb.New(now),
		Status:      "pending",
	}
	if err := s.Keeper.SetGDPRRequest(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for GDPR data request
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_data_requested",
			sdk.NewAttribute("request_id", id),
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("request_type", req.RequestType),
		),
	)

	return &types.MsgRequestGDPRDataResponse{RequestId: id}, nil
}

func (s *msgServer) GenerateTaxReport(goCtx context.Context, req *types.MsgGenerateTaxReport) (*types.MsgGenerateTaxReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.TaxYear == "" || req.Jurisdiction == "" {
		return nil, status.Error(codes.InvalidArgument, "address, tax year, and jurisdiction required")
	}

	// Verify signer - the address requesting the tax report must be the signer
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	requestAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	if !requestAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "address must match transaction signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()
	id := fmt.Sprintf("tax-%s-%s-%d", req.Address, req.TaxYear, now.UnixNano())
	report := &types.TaxReport{
		Id:           id,
		Address:      req.Address,
		TaxYear:      req.TaxYear,
		Jurisdiction: req.Jurisdiction,
		ReportType:   req.ReportType,
		GeneratedAt:  timestamppb.New(now),
		Filed:        false,
	}
	if err := s.Keeper.SetTaxReport(ctx, report); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for tax report generation
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"tax_report_generated",
			sdk.NewAttribute("report_id", id),
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("tax_year", req.TaxYear),
			sdk.NewAttribute("jurisdiction", req.Jurisdiction),
		),
	)

	return &types.MsgGenerateTaxReportResponse{ReportId: id, FilePath: ""}, nil
}

// EraseGDPRData handles GDPR Article 17 "Right to Erasure" requests.
// This function emits an immutable on-chain event recording the erasure request,
// which signals off-chain systems to delete the user's PII. The on-chain commitments
// remain as an audit trail, but the off-chain PII becomes orphaned and unrecoverable.
//
// GDPR compliance:
//   - Article 17 "Right to Erasure": User can request deletion of their personal data
//   - Blockchain immutability: On-chain hashes remain, but PII is deleted off-chain
//   - Audit trail: Erasure request event provides compliance evidence
//   - Irreversibility: Once PII is deleted off-chain, it cannot be recovered
//
// Security considerations:
//   - Only the data subject (address owner) can request erasure
//   - Address must be the transaction signer (authentication)
//   - Erasure event ID is deterministic for audit purposes
//   - Off-chain systems must monitor for erasure events
//
// Implementation note:
//   Off-chain systems (KYC providers, compliance databases) must:
//   1. Monitor the blockchain for "gdpr_data_erased" events
//   2. When event is detected, delete all PII for that address
//   3. Log the erasure in their audit trail
//   4. Confirm deletion to the user (off-chain communication)
func (s *msgServer) EraseGDPRData(goCtx context.Context, req *types.MsgEraseGDPRData) (*types.MsgEraseGDPRDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}

	// Verify signer - only the data subject can request erasure
	signers := req.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	requestAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	if !requestAddr.Equals(signers[0]) {
		return nil, status.Error(codes.PermissionDenied, "only the data subject can request erasure")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()

	// Generate deterministic erasure event ID for audit trail
	erasureEventID := fmt.Sprintf("gdpr-erasure-%s-%d-%d", req.Address, ctx.BlockHeight(), now.Unix())

	// Emit immutable event for off-chain systems to process
	// Off-chain KYC providers and compliance databases must monitor this event
	// and delete all PII associated with this address
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"gdpr_data_erased",
			sdk.NewAttribute("address", req.Address),
			sdk.NewAttribute("erasure_event_id", erasureEventID),
			sdk.NewAttribute("erasure_reason", req.ErasureReason),
			sdk.NewAttribute("erasure_time", now.Format(time.RFC3339)),
			sdk.NewAttribute("block_height", fmt.Sprintf("%d", ctx.BlockHeight())),
		),
	)

	// Note: On-chain commitments/hashes remain for audit purposes
	// but the off-chain PII they reference will be deleted
	// This satisfies GDPR while preserving blockchain immutability

	return &types.MsgEraseGDPRDataResponse{
		Success:         true,
		ErasureEventId:  erasureEventID,
	}, nil
}
