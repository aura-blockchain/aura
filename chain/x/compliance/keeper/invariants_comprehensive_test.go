// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/aequitas/aura/chain/x/compliance/types"
)

type InvariantsComprehensiveTestSuite struct {
	KeeperTestSuite
}

func TestInvariantsComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(InvariantsComprehensiveTestSuite))
}

func (suite *InvariantsComprehensiveTestSuite) TestKYCRecordConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := KYCRecordConsistencyInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)

	expiresAt := time.Now().Add(48 * time.Hour)
	record := &types.KYCRecord{
		Address:       suite.addr("kyc"),
		PiiCommitment: make([]byte, 32),
		KycLevel:      types.KYCLevel_KYC_LEVEL_ADVANCED,
		VerifiedAt:    time.Now(),
		ExpiresAt:     &expiresAt,
	}
	suite.Require().NoError(suite.Keeper.SetKYCRecord(ctx, record))

	msg, broken = inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestSanctionsScreeningInvariant() {
	ctx := suite.SdkCtx
	inv := SanctionsScreeningInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestGDPRDataIntegrityInvariant() {
	ctx := suite.SdkCtx
	inv := GDPRDataIntegrityInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestTaxRecordConsistencyInvariant() {
	ctx := suite.SdkCtx
	inv := TaxRecordConsistencyInvariant(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) TestAllInvariants() {
	ctx := suite.SdkCtx
	inv := AllInvariants(suite.Keeper)

	msg, broken := inv(ctx)
	suite.False(broken)
	suite.Empty(msg)
}

func (suite *InvariantsComprehensiveTestSuite) addr(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	bech32, err := sdk.Bech32ifyAddressBytes("aura", sum[:20])
	suite.Require().NoError(err)
	return bech32
}
