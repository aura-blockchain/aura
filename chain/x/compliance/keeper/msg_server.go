package keeper

import (
	context "context"
	"fmt"
	"sort"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/pkg/log"
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
// The PII data (verification_id, documents, risk_score) must be
// stored off-chain by the KYC provider. Only the cryptographic commitment (SHA-256 hash)
// of the PII is stored on-chain, allowing verification without storing sensitive data.
// Jurisdiction is stored on-chain for OFAC compliance validation.
//
// Security considerations:
//   - Provider must be authorized (checked against params.ApprovedKycProviders)
//   - Provider must be the transaction signer (authentication)
//   - PII commitment must be exactly 32 bytes (SHA-256 hash)
//   - Jurisdiction must be provided and validated against blocked list (OFAC compliance)
//   - Off-chain PII data should be encrypted and stored by the provider
//   - Data erasure requests can be fulfilled off-chain while preserving on-chain audit trail
//
// OFAC compliance:
//   - Jurisdiction is validated against params.BlockedJurisdictions
//   - Users from sanctioned countries (KP, IR, SY, CU, etc.) are rejected
//   - Governance can update the blocked list via params
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
	if req.Jurisdiction == "" {
		return nil, status.Error(codes.InvalidArgument, "jurisdiction is required (ISO 3166-1 alpha-2 country code)")
	}

	// Normalize jurisdiction to uppercase for ISO 3166-1 alpha-2 compliance
	// This allows case-insensitive input while maintaining standard format
	normalizedJurisdiction := strings.ToUpper(strings.TrimSpace(req.Jurisdiction))

	// Validate jurisdiction format (2-letter ISO 3166-1 alpha-2 country code)
	// This validation enforces:
	//   - Exactly 2 uppercase letters (e.g., "US", "GB", "JP")
	//   - No numeric characters
	//   - No special characters
	if err := types.ValidateJurisdictionCode(normalizedJurisdiction); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid jurisdiction: must be 2-letter ISO 3166-1 alpha-2 country code (e.g., 'US', 'GB', 'JP'): %s", err.Error()))
	}

	// Use normalized jurisdiction for all subsequent operations
	req.Jurisdiction = normalizedJurisdiction

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
	log.TxStart(ctx, "MsgSubmitKYC", req.Provider)

	// Check if provider is authorized (validate authority first)
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

	// OFAC compliance: Check if jurisdiction is blocked (sanctioned country)
	// This validation must occur before consent check
	if s.Keeper.IsJurisdictionBlocked(ctx, req.Jurisdiction) {
		return nil, status.Errorf(codes.PermissionDenied,
			"jurisdiction %s is blocked due to OFAC sanctions", req.Jurisdiction)
	}

	// GDPR Consent Enforcement: Verify user has consented to KYC processing (Article 6(1)(a))
	// This must be checked AFTER provider/jurisdiction validation but BEFORE processing user data
	if err := s.Keeper.RequireConsent(ctx, req.Address, "kyc_processing"); err != nil {
		return nil, status.Error(codes.PermissionDenied,
			"user consent required for KYC processing - consent not found or has been withdrawn (GDPR Article 7(3))")
	}

	now := ctx.BlockTime()
	expiresAt := now.Add(time.Duration(params.KycExpiryDays) * 24 * time.Hour)
	record := &types.KYCRecord{
		Address:              req.Address,
		KycLevel:             req.KycLevel,
		Provider:             req.Provider,
		VerifiedAt:           now,
		ExpiresAt:            &expiresAt,
		PiiCommitment:        req.PiiCommitment,
		EnhancedDueDiligence: req.KycLevel == types.KYCLevel_KYC_LEVEL_ADVANCED,
		Jurisdiction:         req.Jurisdiction,
	}

	// Use UpdateKYCRecord for proper version tracking and history preservation
	// This handles deduplication and conflict resolution automatically
	updateReason := "initial_submission"
	if existing, err := s.Keeper.GetKYCRecord(ctx, req.Address); err == nil && existing != nil {
		// Determine update reason based on changes
		if existing.KycLevel != req.KycLevel {
			updateReason = "level_upgrade"
		} else if existing.Provider != req.Provider {
			updateReason = "provider_change"
		} else {
			updateReason = "renewal_or_update"
		}
	}

	if err := s.Keeper.UpdateKYCRecord(ctx, record, updateReason); err != nil {
		log.TxError(ctx, "MsgSubmitKYC", err, "provider", req.Provider, "address", req.Address)
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.TxSuccess(ctx, "MsgSubmitKYC", "provider", req.Provider, "address", req.Address, "kyc_level", req.KycLevel.String(), "jurisdiction", req.Jurisdiction)
	log.StateChange(ctx, "kyc_record", "updated", req.Address)

	// Additional event with jurisdiction and PII commitment (version event emitted by UpdateKYCRecord)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKYCApproved,
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyProvider, req.Provider),
			sdk.NewAttribute(types.AttributeKeyKYCLevel, req.KycLevel.String()),
			sdk.NewAttribute(types.AttributeKeyJurisdiction, req.Jurisdiction),
			sdk.NewAttribute(types.AttributeKeyPIICommitment, fmt.Sprintf("%x", req.PiiCommitment)),
			sdk.NewAttribute("version", fmt.Sprintf("%d", record.Version)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return &types.MsgSubmitKYCResponse{Success: true, Message: fmt.Sprintf("kyc record stored with PII commitment (version %d)", record.Version)}, nil
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

	// GDPR Consent Enforcement: Verify user has consented to AML monitoring (Article 6(1)(a))
	// SAR (Suspicious Activity Report) requires consent for monitoring and analyzing transaction patterns
	if err := s.Keeper.RequireConsent(ctx, req.Address, "aml_monitoring"); err != nil {
		return nil, status.Error(codes.PermissionDenied,
			"user consent required for AML monitoring - consent not found or has been withdrawn (GDPR Article 7(3))")
	}

	now := ctx.BlockTime()
	id := fmt.Sprintf("sar-%s-%d", req.TransactionHash, now.UnixNano())
	nowPtr := now
	activity := &types.SuspiciousActivity{
		Id:              id,
		Address:         req.Address,
		TransactionHash: req.TransactionHash,
		ActivityType:    req.ActivityType,
		Description:     req.Description,
		Amount:          "", // Set externally or via analysis
		DetectedAt:      now,
		ReportedAt:      &nowPtr,
		Indicators:      req.Indicators,
		FiledSar:        false,
	}
	if err := s.Keeper.SetSuspiciousActivity(ctx, activity); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for SAR filing (BSA/AML compliance audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSARReported,
			sdk.NewAttribute(types.AttributeKeyActivityID, id),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyReporter, req.Reporter),
			sdk.NewAttribute(types.AttributeKeyActivityType, req.ActivityType),
			sdk.NewAttribute(types.AttributeKeyTransactionHash, req.TransactionHash),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
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

	// GDPR Consent Enforcement: Verify user has consented to sanctions screening (Article 6(1)(a))
	// Sanctions screening processes user data against OFAC/sanctions lists
	if err := s.Keeper.RequireConsent(ctx, req.Address, "sanctions_screening"); err != nil {
		return nil, status.Error(codes.PermissionDenied,
			"user consent required for sanctions screening - consent not found or has been withdrawn (GDPR Article 7(3))")
	}

	// Check cache with expiry validation (OFAC compliance)
	var result *types.SanctionsScreeningResult
	if !req.ForceRefresh {
		result, err = s.Keeper.GetSanctionsResult(ctx, req.Address)
		if err != nil {
			result = nil
		} else if result != nil {
			// Verify cache has not expired (critical for OFAC compliance)
			// This prevents using stale "CLEAR" status for newly sanctioned addresses
			params := s.Keeper.GetParams(ctx)
			if params.ScreeningCacheHours > 0 {
				cacheAge := ctx.BlockTime().Sub(result.ScreenedAt)
				maxCacheAge := time.Duration(params.ScreeningCacheHours) * time.Hour

				if cacheAge > maxCacheAge {
					// Cache expired - force fresh screening
					result = nil
				}
			}
		}
	}

	// Perform fresh screening if no valid cache exists
	if result == nil {
		result, err = s.performSanctionsScreen(ctx, req.Address)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := s.Keeper.SetSanctionsResult(ctx, result); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// Emit event for sanctions screening (OFAC compliance audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSanctionsScreening,
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyStatus, result.Status.String()),
			sdk.NewAttribute(types.AttributeKeyRequiresReview, fmt.Sprintf("%t", result.RequiresManualReview)),
			sdk.NewAttribute(types.AttributeKeyScreeningResult, result.ScreeningProvider),
			sdk.NewAttribute(types.AttributeKeyMatchCount, fmt.Sprintf("%d", len(result.Matches))),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
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
				if res.ScreenedAt.IsZero() {
					res.ScreenedAt = now
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
		ScreenedAt:           now,
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
		ConsentGivenAt: now,
	}
	if !req.Consented {
		nowPtr := now
		consent.ConsentWithdrawnAt = &nowPtr
	}
	if err := s.Keeper.SetGDPRConsent(ctx, consent); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// GDPR Article 7(3) Enforcement: When consent is withdrawn, immediately
	// restrict data processing and trigger deletion of associated PII
	if !req.Consented {
		// Mark address as "do not process" - this flag must be checked
		// before any data processing operation (enforces Article 7(3))
		if err := s.Keeper.SetProcessingRestriction(ctx, req.Address, true); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to set processing restriction: %s", err.Error()))
		}

		// Trigger data deletion for this consent type
		// This emits an event for off-chain systems (KYC providers, databases)
		// to delete the user's PII associated with this consent type
		if err := s.Keeper.TriggerDataDeletion(ctx, req.Address, req.ConsentType); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to trigger data deletion: %s", err.Error()))
		}

		// Emit withdrawal enforcement event for audit trail
		// (GDPR Article 7(3) compliance - demonstrates that withdrawal has immediate effect)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeGDPRConsentWithdrawn,
				sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
				sdk.NewAttribute(types.AttributeKeyConsentType, req.ConsentType),
				sdk.NewAttribute(types.AttributeKeyProcessingRestricted, "true"),
				sdk.NewAttribute(types.AttributeKeyDeletionTriggered, "true"),
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
				sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	} else {
		// When consent is given, remove processing restriction (if any)
		// This allows data processing to resume for this address
		if err := s.Keeper.SetProcessingRestriction(ctx, req.Address, false); err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to remove processing restriction: %s", err.Error()))
		}

		// Emit standard consent recorded event
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeGDPRConsentRecorded,
				sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
				sdk.NewAttribute(types.AttributeKeyConsentType, req.ConsentType),
				sdk.NewAttribute(types.AttributeKeyConsented, fmt.Sprintf("%t", req.Consented)),
				sdk.NewAttribute(types.AttributeKeyConsentVersion, req.ConsentVersion),
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
				sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
				sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
			),
		)
	}

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
		RequestedAt: now,
		Status:      "pending",
	}
	if err := s.Keeper.SetGDPRRequest(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for GDPR data request (GDPR Article 15 compliance audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGDPRDataRequested,
			sdk.NewAttribute(types.AttributeKeyRequestID, id),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyGDPRRequestType, req.RequestType),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
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

	// Generate default file path if not provided by user
	// If user provides a file path, validate it to prevent path traversal attacks
	filePath := ""
	if req.FilePath != "" {
		// Security: Validate file path to prevent path traversal attacks
		// This prevents attackers from using malicious paths like:
		//   - ../../etc/passwd (directory traversal)
		//   - /etc/passwd (absolute path)
		//   - path/with/<script>injection</script> (XSS/injection)
		if err := types.ValidateFilePath(req.FilePath); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid file path: %v", err)
		}
		filePath = req.FilePath
	}

	report := &types.TaxReport{
		Id:           id,
		Address:      req.Address,
		TaxYear:      req.TaxYear,
		Jurisdiction: req.Jurisdiction,
		ReportType:   req.ReportType,
		GeneratedAt:  now,
		FilePath:     filePath,
		Filed:        false,
	}
	if err := s.Keeper.SetTaxReport(ctx, report); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Emit event for tax report generation (tax compliance audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTaxReportGenerated,
			sdk.NewAttribute(types.AttributeKeyReportID, id),
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyTaxYear, req.TaxYear),
			sdk.NewAttribute(types.AttributeKeyJurisdiction, req.Jurisdiction),
			sdk.NewAttribute(types.AttributeKeyReportType, req.ReportType),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
		),
	)

	return &types.MsgGenerateTaxReportResponse{ReportId: id, FilePath: filePath}, nil
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
	// (GDPR Article 17 "Right to Erasure" compliance audit trail)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeGDPRDataErased,
			sdk.NewAttribute(types.AttributeKeyAddress, req.Address),
			sdk.NewAttribute(types.AttributeKeyErasureEventID, erasureEventID),
			sdk.NewAttribute(types.AttributeKeyErasureReason, req.ErasureReason),
			sdk.NewAttribute(types.AttributeKeyErasureTime, now.Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", ctx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyBlockTime, ctx.BlockTime().Format(time.RFC3339)),
			sdk.NewAttribute(types.AttributeKeyTimestamp, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
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
