package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

type queryServer struct {
	types.UnimplementedQueryServer
	Keeper *Keeper
}

// NewQueryServer wires the compliance keeper into a gRPC query server implementation.
func NewQueryServer(k *Keeper) types.QueryServer {
	return &queryServer{Keeper: k}
}

func (q *queryServer) KycRecord(goCtx context.Context, req *types.QueryKYCRecordRequest) (*types.QueryKYCRecordResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply rate limiting for KYC record queries (may trigger verification)
	// TODO: Re-enable when CheckRateLimit is implemented
	// if err := q.Keeper.CheckRateLimit(ctx, req.Address, "kyc_verification"); err != nil {
	// 	return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", err.Error())
	// }

	record, err := q.Keeper.GetKYCRecord(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &types.QueryKYCRecordResponse{Record: record}, nil
}

func (q *queryServer) AmlProfile(goCtx context.Context, req *types.QueryAMLProfileRequest) (*types.QueryAMLProfileResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply rate limiting for AML profile queries (expensive risk calculation)
	// TODO: Re-enable when CheckRateLimit is implemented
	// if err := q.Keeper.CheckRateLimit(ctx, req.Address, "aml_profile_query"); err != nil {
	// 	return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", err.Error())
	// }

	profile, err := q.Keeper.GetAMLProfile(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &types.QueryAMLProfileResponse{Profile: profile}, nil
}

func (q *queryServer) SanctionsScreening(goCtx context.Context, req *types.QuerySanctionsScreeningRequest) (*types.QuerySanctionsScreeningResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply rate limiting for sanctions screening (expensive external API call)
	if err := q.Keeper.CheckRateLimit(ctx, req.Address, "sanctions_screening"); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", err.Error())
	}

	var result *types.SanctionsScreeningResult
	var err error
	if !req.ForceRefresh {
		result, err = q.Keeper.GetSanctionsResult(ctx, req.Address)
	}
	if req.ForceRefresh || result == nil || err != nil {
		msgSrv := &msgServer{Keeper: q.Keeper}
		result, err = msgSrv.performSanctionsScreen(ctx, req.Address)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if err := q.Keeper.SetSanctionsResult(ctx, result); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &types.QuerySanctionsScreeningResponse{Result: result}, nil
}

func (q *queryServer) TransactionAlerts(goCtx context.Context, req *types.QueryTransactionAlertsRequest) (*types.QueryTransactionAlertsResponse, error) {
	if req == nil || req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply rate limiting for transaction alerts (potentially expensive query)
	if err := q.Keeper.CheckRateLimit(ctx, req.Address, "transaction_alerts"); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", err.Error())
	}

	alerts, err := q.Keeper.GetTransactionAlerts(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if req.UnreviewedOnly {
		filtered := make([]*types.TransactionAlert, 0, len(alerts))
		for _, alert := range alerts {
			if !alert.Reviewed {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}
	return &types.QueryTransactionAlertsResponse{Alerts: alerts}, nil
}

func (q *queryServer) TaxReport(goCtx context.Context, req *types.QueryTaxReportRequest) (*types.QueryTaxReportResponse, error) {
	if req == nil || req.Address == "" || req.TaxYear == "" || req.Jurisdiction == "" {
		return nil, status.Error(codes.InvalidArgument, "address, tax year, and jurisdiction are required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Apply rate limiting for tax report queries (expensive blockchain indexing and calculation)
	if err := q.Keeper.CheckRateLimit(ctx, req.Address, "tax_report_generation"); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", err.Error())
	}

	reports, err := q.Keeper.GetTaxReports(ctx, req.Address)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	for _, report := range reports {
		if report.TaxYear == req.TaxYear && report.Jurisdiction == req.Jurisdiction {
			return &types.QueryTaxReportResponse{Report: report}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "tax report not found")
}
