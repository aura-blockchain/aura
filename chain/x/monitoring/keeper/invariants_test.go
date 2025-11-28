package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	keepertest "github.com/aequitas/aura/chain/testing/testutil/keeper"
	"github.com/aequitas/aura/chain/x/monitoring/types"
)

type InvariantsTestSuite struct {
	suite.Suite

	Keeper *Keeper
	SdkCtx sdk.Context
}

func (suite *InvariantsTestSuite) SetupTest() {
	input := keepertest.CreateTestInput(suite.T())
	suite.Keeper = NewKeeper(input.Cdc, input.StoreKey)
	suite.SdkCtx = input.Ctx

	// Set default params
	params := types.DefaultParams()
	err := suite.Keeper.SetParams(params)
	suite.Require().NoError(err)
}

func TestInvariantsTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsTestSuite))
}

// Test all invariants on empty store
func (suite *InvariantsTestSuite) TestAllInvariantsEmptyStore() {
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass on empty store")
	suite.Empty(msg)
}

// Test invariant registration
func (suite *InvariantsTestSuite) TestRegisterInvariants() {
	registry := sdk.NewInvariantRegistry()
	suite.NotPanics(func() {
		RegisterInvariants(registry, suite.Keeper)
	})
}

// ============================================================================
// ParamsInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestParamsInvariantValid() {
	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "params invariant should pass with default params")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestParamsInvariantInvalid() {
	// Set invalid params (negative retention period)
	params := suite.Keeper.GetParams(suite.SdkCtx)
	params.MetricRetentionSeconds = -100 // Invalid
	err := suite.Keeper.SetParams(params)
	suite.Require().NoError(err)

	inv := ParamsInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "params invariant should fail with invalid params")
	suite.NotEmpty(msg)
}

// ============================================================================
// MetricConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestMetricConsistencyInvariantValid() {
	// Create a valid metric
	metric := &types.Metric{
		Name:       "test_metric",
		MetricType: "counter",
		Value:      100.0,
		Timestamp:  timestamppb.Now(),
		Labels:     map[string]string{"module": "test"},
	}
	err := suite.Keeper.SetMetric(suite.SdkCtx, metric)
	suite.Require().NoError(err)

	inv := MetricConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "metric consistency invariant should pass with valid metric")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestMetricConsistencyInvariantEmptyName() {
	// Create metric with empty name
	metric := &types.Metric{
		Name:       "", // Invalid
		MetricType: "counter",
		Value:      100.0,
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetMetric(suite.SdkCtx, metric)
	suite.Require().NoError(err)

	inv := MetricConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "metric consistency invariant should fail with empty name")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestMetricConsistencyInvariantInvalidType() {
	// Create metric with invalid type
	metric := &types.Metric{
		Name:       "test_metric",
		MetricType: "invalid-type", // Invalid
		Value:      100.0,
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetMetric(suite.SdkCtx, metric)
	suite.Require().NoError(err)

	inv := MetricConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "metric consistency invariant should fail with invalid type")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestMetricConsistencyInvariantNegativeCounter() {
	// Create counter with negative value
	metric := &types.Metric{
		Name:       "test_counter",
		MetricType: "counter",
		Value:      -50.0, // Invalid for counter
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetMetric(suite.SdkCtx, metric)
	suite.Require().NoError(err)

	inv := MetricConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "metric consistency invariant should fail with negative counter value")
	suite.NotEmpty(msg)
}

// ============================================================================
// AlertValidityInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestAlertValidityInvariantValid() {
	// Create a valid alert
	alert := &types.Alert{
		AlertId:     "alert-1",
		Name:        "High CPU Alert",
		Severity:    "warning",
		Status:      "active",
		MetricName:  "cpu_usage",
		TriggeredAt: timestamppb.Now(),
		Description: "CPU usage above threshold",
	}
	err := suite.Keeper.SetAlert(suite.SdkCtx, alert)
	suite.Require().NoError(err)

	inv := AlertValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "alert validity invariant should pass with valid alert")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestAlertValidityInvariantEmptyID() {
	// Create alert with empty ID
	alert := &types.Alert{
		AlertId:     "", // Invalid
		Name:        "Test Alert",
		Severity:    "warning",
		Status:      "active",
		TriggeredAt: timestamppb.Now(),
	}
	err := suite.Keeper.SetAlert(suite.SdkCtx, alert)
	suite.Require().NoError(err)

	inv := AlertValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "alert validity invariant should fail with empty ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestAlertValidityInvariantInvalidSeverity() {
	// Create alert with invalid severity
	alert := &types.Alert{
		AlertId:     "alert-1",
		Name:        "Test Alert",
		Severity:    "invalid-severity", // Invalid
		Status:      "active",
		TriggeredAt: timestamppb.Now(),
	}
	err := suite.Keeper.SetAlert(suite.SdkCtx, alert)
	suite.Require().NoError(err)

	inv := AlertValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "alert validity invariant should fail with invalid severity")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestAlertValidityInvariantResolvedWithoutTime() {
	// Create resolved alert without resolution time
	alert := &types.Alert{
		AlertId:     "alert-1",
		Name:        "Test Alert",
		Severity:    "warning",
		Status:      "resolved",
		TriggeredAt: timestamppb.Now(),
		ResolvedAt:  nil, // Invalid for resolved
	}
	err := suite.Keeper.SetAlert(suite.SdkCtx, alert)
	suite.Require().NoError(err)

	inv := AlertValidityInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "alert validity invariant should fail for resolved alert without resolution time")
	suite.NotEmpty(msg)
}

// ============================================================================
// ThresholdConsistencyInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestThresholdConsistencyInvariantValid() {
	// Create a valid threshold
	threshold := &types.Threshold{
		MetricName:              "cpu_usage",
		Operator:                "gt",
		ThresholdValue:          80.0,
		EvaluationWindowSeconds: 300,
		Severity:                "warning",
	}
	err := suite.Keeper.SetThreshold(suite.SdkCtx, threshold)
	suite.Require().NoError(err)

	inv := ThresholdConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "threshold consistency invariant should pass with valid threshold")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestThresholdConsistencyInvariantEmptyMetric() {
	// Create threshold with empty metric name
	threshold := &types.Threshold{
		MetricName:              "", // Invalid
		Operator:                "gt",
		ThresholdValue:          80.0,
		EvaluationWindowSeconds: 300,
	}
	err := suite.Keeper.SetThreshold(suite.SdkCtx, threshold)
	suite.Require().NoError(err)

	inv := ThresholdConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "threshold consistency invariant should fail with empty metric name")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestThresholdConsistencyInvariantInvalidOperator() {
	// Create threshold with invalid operator
	threshold := &types.Threshold{
		MetricName:              "cpu_usage",
		Operator:                "invalid-op", // Invalid
		ThresholdValue:          80.0,
		EvaluationWindowSeconds: 300,
	}
	err := suite.Keeper.SetThreshold(suite.SdkCtx, threshold)
	suite.Require().NoError(err)

	inv := ThresholdConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "threshold consistency invariant should fail with invalid operator")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestThresholdConsistencyInvariantZeroEvaluationWindow() {
	// Create threshold with zero evaluation window
	threshold := &types.Threshold{
		MetricName:              "cpu_usage",
		Operator:                "gt",
		ThresholdValue:          80.0,
		EvaluationWindowSeconds: 0, // Invalid
	}
	err := suite.Keeper.SetThreshold(suite.SdkCtx, threshold)
	suite.Require().NoError(err)

	inv := ThresholdConsistencyInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "threshold consistency invariant should fail with zero evaluation window")
	suite.NotEmpty(msg)
}

// ============================================================================
// AnomalyDetectionStateInvariant Tests
// ============================================================================

func (suite *InvariantsTestSuite) TestAnomalyDetectionStateInvariantValid() {
	// Create a valid anomaly
	anomaly := &types.Anomaly{
		AnomalyId:           "anomaly-1",
		MetricName:          "response_time",
		ConfidenceScore:     0.85,
		DetectionAlgorithm:  "isolation_forest",
		DetectedAt:          timestamppb.Now(),
		Description:         "Unusual spike in response time",
	}
	err := suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	inv := AnomalyDetectionStateInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "anomaly detection state invariant should pass with valid anomaly")
	suite.Empty(msg)
}

func (suite *InvariantsTestSuite) TestAnomalyDetectionStateInvariantEmptyID() {
	// Create anomaly with empty ID
	anomaly := &types.Anomaly{
		AnomalyId:           "", // Invalid
		MetricName:          "response_time",
		ConfidenceScore:     0.85,
		DetectionAlgorithm:  "isolation_forest",
		DetectedAt:          timestamppb.Now(),
	}
	err := suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	inv := AnomalyDetectionStateInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "anomaly detection state invariant should fail with empty ID")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestAnomalyDetectionStateInvariantInvalidConfidence() {
	// Create anomaly with invalid confidence score (> 1.0)
	anomaly := &types.Anomaly{
		AnomalyId:           "anomaly-1",
		MetricName:          "response_time",
		ConfidenceScore:     1.5, // Invalid (> 1.0)
		DetectionAlgorithm:  "isolation_forest",
		DetectedAt:          timestamppb.Now(),
	}
	err := suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	inv := AnomalyDetectionStateInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "anomaly detection state invariant should fail with confidence > 1.0")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestAnomalyDetectionStateInvariantNegativeConfidence() {
	// Create anomaly with negative confidence score
	anomaly := &types.Anomaly{
		AnomalyId:           "anomaly-1",
		MetricName:          "response_time",
		ConfidenceScore:     -0.5, // Invalid (< 0)
		DetectionAlgorithm:  "isolation_forest",
		DetectedAt:          timestamppb.Now(),
	}
	err := suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	inv := AnomalyDetectionStateInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "anomaly detection state invariant should fail with negative confidence")
	suite.NotEmpty(msg)
}

func (suite *InvariantsTestSuite) TestAnomalyDetectionStateInvariantEmptyAlgorithm() {
	// Create anomaly with empty detection algorithm
	anomaly := &types.Anomaly{
		AnomalyId:           "anomaly-1",
		MetricName:          "response_time",
		ConfidenceScore:     0.85,
		DetectionAlgorithm:  "", // Invalid
		DetectedAt:          timestamppb.Now(),
	}
	err := suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	inv := AnomalyDetectionStateInvariant(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.True(broken, "anomaly detection state invariant should fail with empty algorithm")
	suite.NotEmpty(msg)
}

// ============================================================================
// All Invariants Integration Test
// ============================================================================

func (suite *InvariantsTestSuite) TestAllInvariantsWithValidData() {
	// Setup valid data
	metric := &types.Metric{
		Name:       "cpu_usage",
		MetricType: "gauge",
		Value:      75.5,
		Timestamp:  timestamppb.Now(),
	}
	err := suite.Keeper.SetMetric(suite.SdkCtx, metric)
	suite.Require().NoError(err)

	alert := &types.Alert{
		AlertId:     "alert-1",
		Name:        "High CPU",
		Severity:    "warning",
		Status:      "active",
		MetricName:  "cpu_usage",
		TriggeredAt: timestamppb.Now(),
	}
	err = suite.Keeper.SetAlert(suite.SdkCtx, alert)
	suite.Require().NoError(err)

	threshold := &types.Threshold{
		MetricName:              "cpu_usage",
		Operator:                "gt",
		ThresholdValue:          80.0,
		EvaluationWindowSeconds: 300,
	}
	err = suite.Keeper.SetThreshold(suite.SdkCtx, threshold)
	suite.Require().NoError(err)

	anomaly := &types.Anomaly{
		AnomalyId:           "anomaly-1",
		MetricName:          "cpu_usage",
		ConfidenceScore:     0.75,
		DetectionAlgorithm:  "z_score",
		DetectedAt:          timestamppb.Now(),
	}
	err = suite.Keeper.SetAnomaly(suite.SdkCtx, anomaly)
	suite.Require().NoError(err)

	// Run all invariants
	inv := AllInvariants(suite.Keeper)
	msg, broken := inv(suite.SdkCtx)
	suite.False(broken, "all invariants should pass with valid data")
	suite.Empty(msg)
}
