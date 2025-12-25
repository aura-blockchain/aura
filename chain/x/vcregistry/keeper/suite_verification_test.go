// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestKeeperTestSuiteRuns verifies that KeeperTestSuite is properly initialized
type SuiteVerificationTest struct {
	KeeperTestSuite
}

func TestKeeperTestSuiteRuns(t *testing.T) {
	suite.Run(t, new(SuiteVerificationTest))
}

func (suite *SuiteVerificationTest) TestKeeperInitialized() {
	suite.Require().NotNil(suite.Keeper, "Keeper should be initialized")
	suite.Require().NotNil(suite.SdkCtx, "SdkCtx should be initialized")
	suite.Require().NotNil(suite.Cdc, "Cdc should be initialized")
}

func (suite *SuiteVerificationTest) TestBlockMetadata() {
	suite.Require().Equal(int64(100), suite.GetBlockHeight(), "Block height should be 100")
	suite.Require().NotZero(suite.GetBlockTime(), "Block time should be set")
}
