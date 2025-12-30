// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Session Security Tests (ConfigureSession, SecureEnclave, etc.)
// =============================================================================

type SessionSecurityTestSuite struct {
	KeeperTestSuite
}

func (suite *SessionSecurityTestSuite) TestConfigureSession_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	timeout := &gogotypes.Duration{Seconds: 1800, Nanos: 0} // 30 minutes
	config, err := k.ConfigureSession(ctx, "wallet-1", timeout, true, 300)
	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().NotEmpty(config.SessionId)
	suite.Require().Equal("wallet-1", config.WalletId)
	suite.Require().True(config.AutoLockEnabled)
	suite.Require().Equal(int32(300), config.InactivityThresholdSeconds)
	suite.Require().False(config.Locked)
}

func (suite *SessionSecurityTestSuite) TestConfigureSession_NoAutoLock() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	timeout := &gogotypes.Duration{Seconds: 3600, Nanos: 0}
	config, err := k.ConfigureSession(ctx, "wallet-2", timeout, false, 0)
	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().False(config.AutoLockEnabled)
}

func (suite *SessionSecurityTestSuite) TestStoreInSecureEnclave_SGX() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	encryptedKey := []byte("encrypted-key-material")
	attestation := "attestation-cert-data"

	config, err := k.StoreInSecureEnclave(ctx, "wallet-3", wsproto.EnclaveType_ENCLAVE_TYPE_SGX, encryptedKey, attestation)
	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().NotEmpty(config.EnclaveId)
	suite.Require().Equal("wallet-3", config.WalletId)
	suite.Require().Equal(wsproto.EnclaveType_ENCLAVE_TYPE_SGX, config.EnclaveType)
	suite.Require().Equal(encryptedKey, config.EncryptedKeyMaterial)
	suite.Require().True(config.HardwareBacked)
	suite.Require().Equal(attestation, config.AttestationCertificate)
}

func (suite *SessionSecurityTestSuite) TestStoreInSecureEnclave_TEE() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	config, err := k.StoreInSecureEnclave(ctx, "wallet-4", wsproto.EnclaveType_ENCLAVE_TYPE_TEE, []byte("key"), "cert")
	suite.Require().NoError(err)
	suite.Require().NotNil(config)
	suite.Require().Equal(wsproto.EnclaveType_ENCLAVE_TYPE_TEE, config.EnclaveType)
}

func (suite *SessionSecurityTestSuite) TestCreateEncryptedBackup_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	encryptedSeed := []byte("encrypted-seed-data")
	salt := []byte("random-salt-value")

	backup, err := k.CreateEncryptedBackup(ctx, "wallet-5", encryptedSeed, "AES-256-GCM", "PBKDF2", salt, 100000, wsproto.BackupLocation_BACKUP_LOCATION_CLOUD)
	suite.Require().NoError(err)
	suite.Require().NotNil(backup)
	suite.Require().NotEmpty(backup.BackupId)
	suite.Require().Equal("wallet-5", backup.WalletId)
	suite.Require().Equal(encryptedSeed, backup.EncryptedSeed)
	suite.Require().Equal("AES-256-GCM", backup.EncryptionAlgorithm)
	suite.Require().Equal("PBKDF2", backup.KeyDerivationFunction)
	suite.Require().Equal(salt, backup.Salt)
	suite.Require().Equal(int32(100000), backup.Iterations)
	suite.Require().Equal(wsproto.BackupLocation_BACKUP_LOCATION_CLOUD, backup.Location)
	suite.Require().NotEmpty(backup.Checksum)
}

func (suite *SessionSecurityTestSuite) TestCreateEncryptedBackup_LocalLocation() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	backup, err := k.CreateEncryptedBackup(ctx, "wallet-6", []byte("seed"), "AES-256", "ARGON2", []byte("salt"), 50000, wsproto.BackupLocation_BACKUP_LOCATION_LOCAL)
	suite.Require().NoError(err)
	suite.Require().NotNil(backup)
	suite.Require().Equal(wsproto.BackupLocation_BACKUP_LOCATION_LOCAL, backup.Location)
}

func (suite *SessionSecurityTestSuite) TestConfigureDustFilter_Enabled() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	filter, err := k.ConfigureDustFilter(ctx, "wallet-7", true, "1000", 10, 5)
	suite.Require().NoError(err)
	suite.Require().NotNil(filter)
	suite.Require().Equal("wallet-7", filter.WalletId)
	suite.Require().True(filter.Enabled)
	suite.Require().Equal("1000", filter.MinimumAmount)
	suite.Require().Equal(int32(10), filter.MaxDustTransactionsPerBlock)
	suite.Require().Equal(int32(5), filter.SuspiciousPatternThreshold)
}

func (suite *SessionSecurityTestSuite) TestConfigureDustFilter_Disabled() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	filter, err := k.ConfigureDustFilter(ctx, "wallet-8", false, "0", 0, 0)
	suite.Require().NoError(err)
	suite.Require().NotNil(filter)
	suite.Require().False(filter.Enabled)
}

func (suite *SessionSecurityTestSuite) TestValidateAddressChecksum_Valid() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	isValid, checksum, err := k.ValidateAddressChecksum(ctx, "aura1abcdef123456", wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32)
	suite.Require().NoError(err)
	suite.Require().True(isValid)
	suite.Require().NotEmpty(checksum)
}

func (suite *SessionSecurityTestSuite) TestValidateAddressChecksum_EmptyAddress() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	isValid, checksum, err := k.ValidateAddressChecksum(ctx, "", wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32)
	suite.Require().Error(err)
	suite.Require().False(isValid)
	suite.Require().Empty(checksum)
	suite.Require().Contains(err.Error(), "empty")
}

func (suite *SessionSecurityTestSuite) TestValidateAddressChecksum_DifferentAlgorithms() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// BECH32
	isValid, checksum, err := k.ValidateAddressChecksum(ctx, "aura1test", wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_BECH32)
	suite.Require().NoError(err)
	suite.Require().True(isValid)
	suite.Require().NotEmpty(checksum)

	// EIP55
	isValid2, checksum2, err := k.ValidateAddressChecksum(ctx, "aura1test", wsproto.ChecksumAlgorithm_CHECKSUM_ALGORITHM_EIP55)
	suite.Require().NoError(err)
	suite.Require().True(isValid2)
	suite.Require().NotEmpty(checksum2)
}

func TestSessionSecurityTestSuite(t *testing.T) {
	suite.Run(t, new(SessionSecurityTestSuite))
}
