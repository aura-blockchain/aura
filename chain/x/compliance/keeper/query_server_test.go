package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

func TestQueryKYCRecord(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	record := &types.KYCRecord{Address: "aura1kycquery", KycLevel: types.KYCLevel_KYC_LEVEL_BASIC}
	require.NoError(t, keeper.SetKYCRecord(ctx, record))
	server := NewQueryServer(keeper)
	res, err := server.KycRecord(sdk.WrapSDKContext(ctx), &types.QueryKYCRecordRequest{Address: record.Address})
	require.NoError(t, err)
	require.NotNil(t, res.Record)
	require.Equal(t, record.Address, res.Record.Address)
}

func TestQueryTransactionAlertsFiltersReviewed(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	alertReviewed := &types.TransactionAlert{Id: "reviewed", Reviewed: true}
	alertPending := &types.TransactionAlert{Id: "pending", Reviewed: false}
	require.NoError(t, keeper.AddTransactionAlert(ctx, "aura1alerts", alertReviewed))
	require.NoError(t, keeper.AddTransactionAlert(ctx, "aura1alerts", alertPending))
	server := NewQueryServer(keeper)
	res, err := server.TransactionAlerts(sdk.WrapSDKContext(ctx), &types.QueryTransactionAlertsRequest{Address: "aura1alerts", UnreviewedOnly: true})
	require.NoError(t, err)
	require.Len(t, res.Alerts, 1)
	require.Equal(t, "pending", res.Alerts[0].Id)
}

func TestQueryTaxReportReturnsMatch(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	report := &types.TaxReport{
		Id:           "tax1",
		Address:      "aura1taxquery",
		TaxYear:      "2024",
		Jurisdiction: "US",
		GeneratedAt: time.Now(),
	}
	require.NoError(t, keeper.SetTaxReport(ctx, report))
	server := NewQueryServer(keeper)
	res, err := server.TaxReport(sdk.WrapSDKContext(ctx), &types.QueryTaxReportRequest{Address: report.Address, TaxYear: report.TaxYear, Jurisdiction: report.Jurisdiction})
	require.NoError(t, err)
	require.Equal(t, report.TaxYear, res.Report.TaxYear)
}

func TestQueryParams(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewQueryServer(keeper)

	// Test successful params query
	res, err := server.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Params)

	// Test nil request
	_, err = server.Params(sdk.WrapSDKContext(ctx), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid request")
}

func TestQueryKycHistory(t *testing.T) {
	keeper, ctx := setupTestKeeper(t)
	server := NewQueryServer(keeper)

	addr := "aura1historytest"

	// Add KYC history entries
	history1 := &types.KYCHistoryEntry{
		Address:   addr,
		Status:    types.KYCStatus_APPROVED,
		Timestamp: time.Now().Add(-2 * time.Hour),
		Reason:    "Initial approval",
	}
	history2 := &types.KYCHistoryEntry{
		Address:   addr,
		Status:    types.KYCStatus_EXPIRED,
		Timestamp: time.Now().Add(-1 * time.Hour),
		Reason:    "Expired",
	}

	err := keeper.AddKYCHistory(ctx, addr, history1)
	require.NoError(t, err)
	err = keeper.AddKYCHistory(ctx, addr, history2)
	require.NoError(t, err)

	// Test successful history query
	res, err := server.KycHistory(sdk.WrapSDKContext(ctx), &types.QueryKYCHistoryRequest{
		Address: addr,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.History, 2)

	// Test nil request
	_, err = server.KycHistory(sdk.WrapSDKContext(ctx), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid request")

	// Test empty address
	_, err = server.KycHistory(sdk.WrapSDKContext(ctx), &types.QueryKYCHistoryRequest{
		Address: "",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "address cannot be empty")
}
