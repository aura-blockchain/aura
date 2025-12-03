package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

// AllKYCRecords returns all KYC records with pagination support.
// This implements DoS protection by limiting the number of records per query.
//
// Security considerations:
//   - Default limit is 100 if not specified (prevents unbounded queries)
//   - Maximum limit of 1000 enforced by Cosmos SDK query.Paginate
//   - Efficient pagination using store iterator with next key
//
// Example usage:
//   aurad query compliance all-kyc-records --limit 50
//   aurad query compliance all-kyc-records --page-key <next-key>
func (q *queryServer) AllKYCRecords(goCtx context.Context, req *types.QueryAllKYCRecordsRequest) (*types.QueryAllKYCRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	records, pageRes, err := q.Keeper.GetAllKYCRecordsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllKYCRecordsResponse{
		Records:    records,
		Pagination: pageRes,
	}, nil
}

// AllAMLProfiles returns all AML profiles with pagination support.
func (q *queryServer) AllAMLProfiles(goCtx context.Context, req *types.QueryAllAMLProfilesRequest) (*types.QueryAllAMLProfilesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	profiles, pageRes, err := q.Keeper.GetAllAMLProfilesPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAMLProfilesResponse{
		Profiles:   profiles,
		Pagination: pageRes,
	}, nil
}

// AllSanctionsResults returns all sanctions screening results with pagination support.
func (q *queryServer) AllSanctionsResults(goCtx context.Context, req *types.QueryAllSanctionsResultsRequest) (*types.QueryAllSanctionsResultsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	results, pageRes, err := q.Keeper.GetAllSanctionsResultsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSanctionsResultsResponse{
		Results:    results,
		Pagination: pageRes,
	}, nil
}

// AllTransactionAlerts returns all transaction alerts across all addresses with pagination support.
func (q *queryServer) AllTransactionAlerts(goCtx context.Context, req *types.QueryAllTransactionAlertsRequest) (*types.QueryAllTransactionAlertsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	alerts, pageRes, err := q.Keeper.GetAllTransactionAlertsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert map to list for response
	var alertLists []*types.TransactionAlertList
	for address, alertList := range alerts {
		// Store address in first alert if list is not empty (for reference)
		// In production, you might want a dedicated wrapper type
		alertLists = append(alertLists, &types.TransactionAlertList{
			Alerts: alertList,
		})
		_ = address // Address is implicit in the alert.Address field
	}

	return &types.QueryAllTransactionAlertsResponse{
		Alerts:     alertLists,
		Pagination: pageRes,
	}, nil
}

// AllGDPRConsents returns all GDPR consents across all addresses with pagination support.
func (q *queryServer) AllGDPRConsents(goCtx context.Context, req *types.QueryAllGDPRConsentsRequest) (*types.QueryAllGDPRConsentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	consents, pageRes, err := q.Keeper.GetAllGDPRConsentsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert map to list for response
	var consentLists []*types.GDPRConsentList
	for _, consentList := range consents {
		consentLists = append(consentLists, &types.GDPRConsentList{
			Consents: consentList,
		})
	}

	return &types.QueryAllGDPRConsentsResponse{
		Consents:   consentLists,
		Pagination: pageRes,
	}, nil
}

// AllTaxReports returns all tax reports across all addresses with pagination support.
func (q *queryServer) AllTaxReports(goCtx context.Context, req *types.QueryAllTaxReportsRequest) (*types.QueryAllTaxReportsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	reports, pageRes, err := q.Keeper.GetAllTaxReportsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert map to list for response
	var reportLists []*types.TaxReportList
	for _, reportList := range reports {
		reportLists = append(reportLists, &types.TaxReportList{
			Reports: reportList,
		})
	}

	return &types.QueryAllTaxReportsResponse{
		Reports:    reportLists,
		Pagination: pageRes,
	}, nil
}

// AllGDPRRequests returns all GDPR data requests with pagination support.
func (q *queryServer) AllGDPRRequests(goCtx context.Context, req *types.QueryAllGDPRRequestsRequest) (*types.QueryAllGDPRRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requests, pageRes, err := q.Keeper.GetAllGDPRRequestsPaginated(ctx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllGDPRRequestsResponse{
		Requests:   requests,
		Pagination: pageRes,
	}, nil
}
