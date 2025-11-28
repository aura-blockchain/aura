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

func (s *msgServer) SubmitKYC(goCtx context.Context, req *types.MsgSubmitKYC) (*types.MsgSubmitKYCResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	now := ctx.BlockTime()
	params := s.Keeper.GetParams(ctx)
	expiresAt := timestamppb.New(now.Add(time.Duration(params.KycExpiryDays) * 24 * time.Hour))
	record := &types.KYCRecord{
		Address:              req.Address,
		KycLevel:             req.KycLevel,
		Provider:             req.Provider,
		VerificationId:       req.VerificationId,
		Documents:            req.Documents,
		Jurisdiction:         req.Jurisdiction,
		VerifiedAt:           timestamppb.New(now),
		ExpiresAt:            expiresAt,
		EnhancedDueDiligence: req.KycLevel == types.KYCLevel_KYC_LEVEL_ADVANCED,
	}
	if err := s.Keeper.SetKYCRecord(ctx, record); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.MsgSubmitKYCResponse{Success: true, Message: "kyc record stored"}, nil
}

func (s *msgServer) ReportSuspiciousActivity(goCtx context.Context, req *types.MsgReportSuspiciousActivity) (*types.MsgReportSuspiciousActivityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.TransactionHash == "" {
		return nil, status.Error(codes.InvalidArgument, "address and transaction hash required")
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
	return &types.MsgReportSuspiciousActivityResponse{ActivityId: id}, nil
}

func (s *msgServer) ScreenSanctions(goCtx context.Context, req *types.MsgScreenSanctions) (*types.MsgScreenSanctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	var result *types.SanctionsScreeningResult
	var err error
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
	return &types.MsgRecordGDPRConsentResponse{Success: true}, nil
}

func (s *msgServer) RequestGDPRData(goCtx context.Context, req *types.MsgRequestGDPRData) (*types.MsgRequestGDPRDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.RequestType == "" {
		return nil, status.Error(codes.InvalidArgument, "address and request type required")
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
	return &types.MsgRequestGDPRDataResponse{RequestId: id}, nil
}

func (s *msgServer) GenerateTaxReport(goCtx context.Context, req *types.MsgGenerateTaxReport) (*types.MsgGenerateTaxReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.Address == "" || req.TaxYear == "" || req.Jurisdiction == "" {
		return nil, status.Error(codes.InvalidArgument, "address, tax year, and jurisdiction required")
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
	return &types.MsgGenerateTaxReportResponse{ReportId: id, FilePath: ""}, nil
}
