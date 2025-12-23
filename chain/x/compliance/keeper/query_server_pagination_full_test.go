package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/keeper"
	"github.com/aequitas/aura/chain/x/compliance/types"
)

type QueryServerPaginationTestSuite struct {
	suite.Suite
	*keeper.TestSuite
	queryServer types.QueryServer
}

func TestQueryServerPaginationTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerPaginationTestSuite))
}

func (suite *QueryServerPaginationTestSuite) SetupTest() {
	suite.TestSuite = keeper.NewTestSuite()
	suite.TestSuite.SetupTest()
	suite.queryServer = keeper.NewQueryServer(suite.Keeper)
}

// ============================================================================
// AllKYCRecords Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllKYCRecords_Success() {
	// Create multiple KYC records
	for i := 0; i < 5; i++ {
		record := &types.KYCRecord{
			Address:      suite.GenAddress(i),
			Status:       types.KYCStatus_APPROVED,
			KycLevel:     1,
			Jurisdiction: "US",
			Provider:     "test-provider",
			SubmittedAt:  time.Now(),
		}
		err := suite.Keeper.SetKYCRecord(suite.Ctx, record)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllKYCRecords(suite.Ctx, &types.QueryAllKYCRecordsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.Records, 5)
	suite.Require().NotNil(resp.Pagination)
}

func (suite *QueryServerPaginationTestSuite) TestAllKYCRecords_WithPagination() {
	// Create 15 records
	for i := 0; i < 15; i++ {
		record := &types.KYCRecord{
			Address:      suite.GenAddress(i),
			Status:       types.KYCStatus_APPROVED,
			KycLevel:     1,
			Jurisdiction: "US",
			Provider:     "test-provider",
			SubmittedAt:  time.Now(),
		}
		err := suite.Keeper.SetKYCRecord(suite.Ctx, record)
		suite.Require().NoError(err)
	}

	// First page
	resp1, err := suite.queryServer.AllKYCRecords(suite.Ctx, &types.QueryAllKYCRecordsRequest{
		Pagination: &query.PageRequest{
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp1.Records, 5)

	// Second page
	resp2, err := suite.queryServer.AllKYCRecords(suite.Ctx, &types.QueryAllKYCRecordsRequest{
		Pagination: &query.PageRequest{
			Key:   resp1.Pagination.NextKey,
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp2.Records, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllKYCRecords_NilRequest() {
	_, err := suite.queryServer.AllKYCRecords(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

func (suite *QueryServerPaginationTestSuite) TestAllKYCRecords_Empty() {
	resp, err := suite.queryServer.AllKYCRecords(suite.Ctx, &types.QueryAllKYCRecordsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Empty(resp.Records)
}

// ============================================================================
// AllAMLProfiles Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllAMLProfiles_Success() {
	// Create multiple AML profiles
	for i := 0; i < 5; i++ {
		profile := &types.AMLProfile{
			Address:           suite.GenAddress(i),
			RiskLevel:         types.RiskLevel_LOW,
			TotalTransactions: uint64(i * 10),
			TotalVolume:       sdkmath.NewInt(int64(i * 1000)),
			LastAssessment:    time.Now(),
		}
		err := suite.Keeper.SetAMLProfile(suite.Ctx, profile)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllAMLProfiles(suite.Ctx, &types.QueryAllAMLProfilesRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.Profiles, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllAMLProfiles_WithPagination() {
	// Create 12 profiles
	for i := 0; i < 12; i++ {
		profile := &types.AMLProfile{
			Address:           suite.GenAddress(i),
			RiskLevel:         types.RiskLevel_MEDIUM,
			TotalTransactions: uint64(i),
			TotalVolume:       sdkmath.NewInt(int64(i * 500)),
			LastAssessment:    time.Now(),
		}
		err := suite.Keeper.SetAMLProfile(suite.Ctx, profile)
		suite.Require().NoError(err)
	}

	// First page
	resp1, err := suite.queryServer.AllAMLProfiles(suite.Ctx, &types.QueryAllAMLProfilesRequest{
		Pagination: &query.PageRequest{
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp1.Profiles, 5)

	// Second page
	resp2, err := suite.queryServer.AllAMLProfiles(suite.Ctx, &types.QueryAllAMLProfilesRequest{
		Pagination: &query.PageRequest{
			Key:   resp1.Pagination.NextKey,
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp2.Profiles, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllAMLProfiles_NilRequest() {
	_, err := suite.queryServer.AllAMLProfiles(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

// ============================================================================
// AllSanctionsResults Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllSanctionsResults_Success() {
	// Create multiple sanctions results
	for i := 0; i < 5; i++ {
		result := &types.SanctionsResult{
			Address:    suite.GenAddress(i),
			Flagged:    false,
			Reason:     "clear",
			CheckedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		err := suite.Keeper.SetSanctionsResult(suite.Ctx, result)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllSanctionsResults(suite.Ctx, &types.QueryAllSanctionsResultsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.Results, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllSanctionsResults_NilRequest() {
	_, err := suite.queryServer.AllSanctionsResults(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

// ============================================================================
// AllTransactionAlerts Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllTransactionAlerts_Success() {
	// Create multiple transaction alerts
	for i := 0; i < 3; i++ {
		addr := suite.GenAddress(i)
		alert := &types.TransactionAlert{
			Address:         addr,
			TransactionHash: "hash" + string(rune(i)),
			AlertType:       "velocity",
			Severity:        types.AlertSeverity_MEDIUM,
			Amount:          sdkmath.NewInt(10000),
			Triggered:       time.Now(),
			Message:         "Alert message",
		}
		err := suite.Keeper.AddTransactionAlert(suite.Ctx, addr, alert)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllTransactionAlerts(suite.Ctx, &types.QueryAllTransactionAlertsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.Alerts)
}

func (suite *QueryServerPaginationTestSuite) TestAllTransactionAlerts_NilRequest() {
	_, err := suite.queryServer.AllTransactionAlerts(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

func (suite *QueryServerPaginationTestSuite) TestAllTransactionAlerts_Empty() {
	resp, err := suite.queryServer.AllTransactionAlerts(suite.Ctx, &types.QueryAllTransactionAlertsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Empty(resp.Alerts)
}

// ============================================================================
// AllGDPRConsents Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllGDPRConsents_Success() {
	// Create multiple GDPR consents
	for i := 0; i < 3; i++ {
		consent := &types.GDPRConsent{
			Address:        suite.GenAddress(i),
			ConsentType:    "data_processing",
			Consented:      true,
			ConsentGivenAt: time.Now(),
			ConsentVersion: "v1.0",
		}
		err := suite.Keeper.SetGDPRConsent(suite.Ctx, consent)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllGDPRConsents(suite.Ctx, &types.QueryAllGDPRConsentsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.Consents)
}

func (suite *QueryServerPaginationTestSuite) TestAllGDPRConsents_NilRequest() {
	_, err := suite.queryServer.AllGDPRConsents(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

func (suite *QueryServerPaginationTestSuite) TestAllGDPRConsents_Empty() {
	resp, err := suite.queryServer.AllGDPRConsents(suite.Ctx, &types.QueryAllGDPRConsentsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Empty(resp.Consents)
}

// ============================================================================
// AllTaxReports Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllTaxReports_Success() {
	// Create multiple tax reports
	for i := 0; i < 3; i++ {
		addr := suite.GenAddress(i)
		report := &types.TaxReport{
			Address:       addr,
			TaxYear:       2024,
			Jurisdiction:  "US",
			TotalIncome:   sdkmath.NewInt(100000),
			TotalGains:    sdkmath.NewInt(10000),
			TotalLosses:   sdkmath.NewInt(5000),
			GeneratedAt:   time.Now(),
			ReportVersion: "v1.0",
		}
		reports := []*types.TaxReport{report}
		err := suite.Keeper.SetTaxReport(suite.Ctx, addr, 2024, "US", reports)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllTaxReports(suite.Ctx, &types.QueryAllTaxReportsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().NotEmpty(resp.Reports)
}

func (suite *QueryServerPaginationTestSuite) TestAllTaxReports_NilRequest() {
	_, err := suite.queryServer.AllTaxReports(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

func (suite *QueryServerPaginationTestSuite) TestAllTaxReports_Empty() {
	resp, err := suite.queryServer.AllTaxReports(suite.Ctx, &types.QueryAllTaxReportsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Empty(resp.Reports)
}

// ============================================================================
// AllGDPRRequests Tests
// ============================================================================

func (suite *QueryServerPaginationTestSuite) TestAllGDPRRequests_Success() {
	// Create multiple GDPR requests
	for i := 0; i < 5; i++ {
		request := &types.GDPRRequest{
			Id:          string(rune('a' + i)),
			Address:     suite.GenAddress(i),
			RequestType: types.GDPRRequestType_ACCESS,
			RequestedAt: time.Now(),
			Status:      types.GDPRStatus_PENDING,
		}
		err := suite.Keeper.SetGDPRRequest(suite.Ctx, request)
		suite.Require().NoError(err)
	}

	resp, err := suite.queryServer.AllGDPRRequests(suite.Ctx, &types.QueryAllGDPRRequestsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Len(resp.Requests, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllGDPRRequests_WithPagination() {
	// Create 10 requests
	for i := 0; i < 10; i++ {
		request := &types.GDPRRequest{
			Id:          string(rune('a' + i)),
			Address:     suite.GenAddress(i),
			RequestType: types.GDPRRequestType_ERASURE,
			RequestedAt: time.Now(),
			Status:      types.GDPRStatus_PENDING,
		}
		err := suite.Keeper.SetGDPRRequest(suite.Ctx, request)
		suite.Require().NoError(err)
	}

	// First page
	resp1, err := suite.queryServer.AllGDPRRequests(suite.Ctx, &types.QueryAllGDPRRequestsRequest{
		Pagination: &query.PageRequest{
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp1.Requests, 5)

	// Second page
	resp2, err := suite.queryServer.AllGDPRRequests(suite.Ctx, &types.QueryAllGDPRRequestsRequest{
		Pagination: &query.PageRequest{
			Key:   resp1.Pagination.NextKey,
			Limit: 5,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Len(resp2.Requests, 5)
}

func (suite *QueryServerPaginationTestSuite) TestAllGDPRRequests_NilRequest() {
	_, err := suite.queryServer.AllGDPRRequests(suite.Ctx, nil)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid request")
}

func (suite *QueryServerPaginationTestSuite) TestAllGDPRRequests_Empty() {
	resp, err := suite.queryServer.AllGDPRRequests(suite.Ctx, &types.QueryAllGDPRRequestsRequest{
		Pagination: &query.PageRequest{
			Limit: 10,
		},
	})

	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
	suite.Require().Empty(resp.Requests)
}
