// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// =============================================================================
// Device Fingerprinting Tests
// =============================================================================

type DeviceFingerprintTestSuite struct {
	KeeperTestSuite
}

func (suite *DeviceFingerprintTestSuite) TestRegisterDevice() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("unique-fingerprint-data")

	device, err := k.RegisterDevice(ctx, "wallet-1", "device-1", "My Phone", "mobile", fingerprint)
	suite.Require().NoError(err)
	suite.Require().NotNil(device)
	suite.Require().Equal("device-1", device.DeviceId)
	suite.Require().Equal("wallet-1", device.WalletId)
	suite.Require().Equal("My Phone", device.DeviceName)
	suite.Require().Equal("mobile", device.DeviceType)
	suite.Require().True(device.Trusted)
	suite.Require().NotEmpty(device.FingerprintHash)
}

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("unique-fingerprint-data")

	// Register device first
	_, err := k.RegisterDevice(ctx, "wallet-2", "device-2", "My Laptop", "desktop", fingerprint)
	suite.Require().NoError(err)

	// Verify with same fingerprint
	verified, err := k.VerifyDevice(ctx, "wallet-2", "device-2", fingerprint)
	suite.Require().NoError(err)
	suite.Require().True(verified)
}

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_WrongFingerprint() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("original-fingerprint")
	wrongFingerprint := []byte("different-fingerprint")

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-3", "device-3", "Device", "mobile", fingerprint)
	suite.Require().NoError(err)

	// Verify with wrong fingerprint
	verified, err := k.VerifyDevice(ctx, "wallet-3", "device-3", wrongFingerprint)
	suite.Require().Error(err)
	suite.Require().False(verified)
	suite.Require().Contains(err.Error(), "fingerprint mismatch")
}

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_NotRegistered() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("some-fingerprint")

	// Try to verify non-existent device
	verified, err := k.VerifyDevice(ctx, "wallet-x", "nonexistent", fingerprint)
	suite.Require().Error(err)
	suite.Require().False(verified)
	suite.Require().Contains(err.Error(), "not registered")
}

func (suite *DeviceFingerprintTestSuite) TestRevokeDevice_Success() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("fingerprint-to-revoke")

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-4", "device-4", "My Phone", "mobile", fingerprint)
	suite.Require().NoError(err)

	// Revoke device
	err = k.RevokeDevice(ctx, "wallet-4", "device-4")
	suite.Require().NoError(err)

	// Verify should now fail (device not trusted)
	verified, err := k.VerifyDevice(ctx, "wallet-4", "device-4", fingerprint)
	suite.Require().Error(err)
	suite.Require().False(verified)
	suite.Require().Contains(err.Error(), "not trusted")
}

func (suite *DeviceFingerprintTestSuite) TestRevokeDevice_NotFound() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	err := k.RevokeDevice(ctx, "wallet-x", "nonexistent-device")
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "not found")
}

func (suite *DeviceFingerprintTestSuite) TestGetDevices_Empty() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	devices, err := k.GetDevices(ctx, "wallet-no-devices")
	suite.Require().NoError(err)
	// May return nil or empty slice
	suite.Require().True(len(devices) == 0)
}

func (suite *DeviceFingerprintTestSuite) TestGetDevices_WithDevices() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register multiple devices for same wallet
	_, err := k.RegisterDevice(ctx, "wallet-5", "device-a", "Phone", "mobile", []byte("fp-a"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-5", "device-b", "Laptop", "desktop", []byte("fp-b"))
	suite.Require().NoError(err)

	devices, err := k.GetDevices(ctx, "wallet-5")
	suite.Require().NoError(err)
	suite.Require().Len(devices, 2)
}

func (suite *DeviceFingerprintTestSuite) TestDetectAnomalousDevice_NewDevice() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register a known device
	_, err := k.RegisterDevice(ctx, "wallet-6", "known-device", "Known Phone", "mobile", []byte("known-fp"))
	suite.Require().NoError(err)

	// Check for anomalous (unknown) device
	isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-6", "unknown-device")
	suite.Require().NoError(err)
	suite.Require().True(isAnomalous)
}

func (suite *DeviceFingerprintTestSuite) TestDetectAnomalousDevice_KnownDevice() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register a known device
	_, err := k.RegisterDevice(ctx, "wallet-7", "my-device", "My Phone", "mobile", []byte("my-fp"))
	suite.Require().NoError(err)

	// Check for known device - should not be anomalous
	isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-7", "my-device")
	suite.Require().NoError(err)
	suite.Require().False(isAnomalous)
}

func (suite *DeviceFingerprintTestSuite) TestHashFingerprint() {
	k := suite.GetKeeper()

	fingerprint := []byte("test-fingerprint")
	hash := k.hashFingerprint(fingerprint)

	suite.Require().NotEmpty(hash)
	suite.Require().Len(hash, 32) // SHA-256 produces 32 bytes

	// Same input should produce same hash
	hash2 := k.hashFingerprint(fingerprint)
	suite.Require().Equal(hash, hash2)

	// Different input should produce different hash
	hash3 := k.hashFingerprint([]byte("different"))
	suite.Require().NotEqual(hash, hash3)
}

func TestDeviceFingerprintTestSuite(t *testing.T) {
	suite.Run(t, new(DeviceFingerprintTestSuite))
}

// =============================================================================
// Additional Device Registration Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestRegisterDevice_EmptyFingerprint() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Empty fingerprint should still work
	device, err := k.RegisterDevice(ctx, "wallet-emptyFP", "device-emptyFP", "Empty FP Device", "mobile", []byte{})
	suite.Require().NoError(err)
	suite.Require().NotNil(device)
	suite.Require().Empty(device.Fingerprint)
	suite.Require().NotEmpty(device.FingerprintHash) // Hash of empty still produces hash
}

func (suite *DeviceFingerprintTestSuite) TestRegisterDevice_UpdateExisting() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint1 := []byte("fingerprint-v1")
	fingerprint2 := []byte("fingerprint-v2")

	// Register device first
	device1, err := k.RegisterDevice(ctx, "wallet-update", "device-update", "Device V1", "mobile", fingerprint1)
	suite.Require().NoError(err)
	suite.Require().Equal("Device V1", device1.DeviceName)

	// Register again with different data (overwrite)
	device2, err := k.RegisterDevice(ctx, "wallet-update", "device-update", "Device V2", "tablet", fingerprint2)
	suite.Require().NoError(err)
	suite.Require().Equal("Device V2", device2.DeviceName)
	suite.Require().Equal("tablet", device2.DeviceType)

	// Verify only one device exists
	devices, err := k.GetDevices(ctx, "wallet-update")
	suite.Require().NoError(err)
	suite.Require().Len(devices, 1)
	suite.Require().Equal("Device V2", devices[0].DeviceName)
}

func (suite *DeviceFingerprintTestSuite) TestRegisterDevice_MultipleWallets() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register same device ID for different wallets
	_, err := k.RegisterDevice(ctx, "wallet-A", "shared-device", "A's Device", "mobile", []byte("fp-a"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-B", "shared-device", "B's Device", "mobile", []byte("fp-b"))
	suite.Require().NoError(err)

	// Each wallet should have its own device
	devicesA, err := k.GetDevices(ctx, "wallet-A")
	suite.Require().NoError(err)
	suite.Require().Len(devicesA, 1)
	suite.Require().Equal("A's Device", devicesA[0].DeviceName)

	devicesB, err := k.GetDevices(ctx, "wallet-B")
	suite.Require().NoError(err)
	suite.Require().Len(devicesB, 1)
	suite.Require().Equal("B's Device", devicesB[0].DeviceName)
}

func (suite *DeviceFingerprintTestSuite) TestRegisterDevice_LargeFingerprint() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Large fingerprint (1MB)
	largeFingerprint := make([]byte, 1024*1024)
	for i := range largeFingerprint {
		largeFingerprint[i] = byte(i % 256)
	}

	device, err := k.RegisterDevice(ctx, "wallet-large-fp", "device-large", "Large FP", "desktop", largeFingerprint)
	suite.Require().NoError(err)
	suite.Require().NotNil(device)
	suite.Require().Len(device.FingerprintHash, 32)
}

// =============================================================================
// Additional Device Verification Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_AfterRevoke() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("revoke-test-fingerprint")

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-revoke-verify", "device-rv", "Device", "mobile", fingerprint)
	suite.Require().NoError(err)

	// Verify works before revoke
	verified, err := k.VerifyDevice(ctx, "wallet-revoke-verify", "device-rv", fingerprint)
	suite.Require().NoError(err)
	suite.Require().True(verified)

	// Revoke device
	err = k.RevokeDevice(ctx, "wallet-revoke-verify", "device-rv")
	suite.Require().NoError(err)

	// Verify should fail after revoke
	verified, err = k.VerifyDevice(ctx, "wallet-revoke-verify", "device-rv", fingerprint)
	suite.Require().Error(err)
	suite.Require().False(verified)
	suite.Require().Contains(err.Error(), "not trusted")
}

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_UpdatesLastSeen() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	fingerprint := []byte("lastseen-fingerprint")

	// Register device
	device, err := k.RegisterDevice(ctx, "wallet-lastseen", "device-ls", "Device", "mobile", fingerprint)
	suite.Require().NoError(err)
	initialLastSeen := device.LastSeenAt

	// Verify device
	verified, err := k.VerifyDevice(ctx, "wallet-lastseen", "device-ls", fingerprint)
	suite.Require().NoError(err)
	suite.Require().True(verified)

	// Get device and check last seen updated (or same if same block)
	devices, err := k.GetDevices(ctx, "wallet-lastseen")
	suite.Require().NoError(err)
	suite.Require().Len(devices, 1)
	// LastSeenAt should be >= initial (same or updated)
	suite.Require().GreaterOrEqual(devices[0].LastSeenAt.Seconds, initialLastSeen.Seconds)
}

func (suite *DeviceFingerprintTestSuite) TestVerifyDevice_EmptyFingerprint() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register with empty fingerprint
	_, err := k.RegisterDevice(ctx, "wallet-empty-verify", "device-empty", "Empty", "mobile", []byte{})
	suite.Require().NoError(err)

	// Verify with empty fingerprint
	verified, err := k.VerifyDevice(ctx, "wallet-empty-verify", "device-empty", []byte{})
	suite.Require().NoError(err)
	suite.Require().True(verified)

	// Verify with non-empty fingerprint should fail
	verified, err = k.VerifyDevice(ctx, "wallet-empty-verify", "device-empty", []byte("not-empty"))
	suite.Require().Error(err)
	suite.Require().False(verified)
}

// =============================================================================
// DetectAnomalousDevice Extended Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestDetectAnomalousDevice_RevokedDevice() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-anom-revoke", "device-ar", "Device", "mobile", []byte("fp-ar"))
	suite.Require().NoError(err)

	// Should not be anomalous initially
	isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-anom-revoke", "device-ar")
	suite.Require().NoError(err)
	suite.Require().False(isAnomalous)

	// Revoke device
	err = k.RevokeDevice(ctx, "wallet-anom-revoke", "device-ar")
	suite.Require().NoError(err)

	// Should be anomalous after revoke (not trusted)
	isAnomalous, err = k.DetectAnomalousDevice(ctx, "wallet-anom-revoke", "device-ar")
	suite.Require().NoError(err)
	suite.Require().True(isAnomalous)
}

func (suite *DeviceFingerprintTestSuite) TestDetectAnomalousDevice_NoDevices() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Wallet with no devices, check any device
	isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-no-devices", "any-device")
	suite.Require().NoError(err)
	suite.Require().True(isAnomalous) // Unknown device is anomalous
}

func (suite *DeviceFingerprintTestSuite) TestDetectAnomalousDevice_MultipleDevices() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register multiple devices
	_, err := k.RegisterDevice(ctx, "wallet-multi-anom", "device-1", "D1", "mobile", []byte("fp-1"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-multi-anom", "device-2", "D2", "desktop", []byte("fp-2"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-multi-anom", "device-3", "D3", "tablet", []byte("fp-3"))
	suite.Require().NoError(err)

	// All known devices should not be anomalous
	for _, deviceID := range []string{"device-1", "device-2", "device-3"} {
		isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-multi-anom", deviceID)
		suite.Require().NoError(err)
		suite.Require().False(isAnomalous, "device %s should not be anomalous", deviceID)
	}

	// Unknown device should be anomalous
	isAnomalous, err := k.DetectAnomalousDevice(ctx, "wallet-multi-anom", "device-unknown")
	suite.Require().NoError(err)
	suite.Require().True(isAnomalous)
}

// =============================================================================
// GetDevices Extended Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestGetDevices_ManyDevices() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	walletID := "wallet-many-devices"

	// Register 20 devices
	for i := 1; i <= 20; i++ {
		deviceID := "device-" + string(rune('a'+i-1))
		_, err := k.RegisterDevice(ctx, walletID, deviceID, "Device "+deviceID, "mobile", []byte("fp-"+deviceID))
		suite.Require().NoError(err)
	}

	devices, err := k.GetDevices(ctx, walletID)
	suite.Require().NoError(err)
	suite.Require().Len(devices, 20)
}

func (suite *DeviceFingerprintTestSuite) TestGetDevices_IsolationBetweenWallets() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register devices for wallet-iso-1
	_, err := k.RegisterDevice(ctx, "wallet-iso-1", "d1", "D1", "mobile", []byte("fp1"))
	suite.Require().NoError(err)
	_, err = k.RegisterDevice(ctx, "wallet-iso-1", "d2", "D2", "desktop", []byte("fp2"))
	suite.Require().NoError(err)

	// Register devices for wallet-iso-2
	_, err = k.RegisterDevice(ctx, "wallet-iso-2", "d3", "D3", "tablet", []byte("fp3"))
	suite.Require().NoError(err)

	// Verify isolation
	devices1, err := k.GetDevices(ctx, "wallet-iso-1")
	suite.Require().NoError(err)
	suite.Require().Len(devices1, 2)

	devices2, err := k.GetDevices(ctx, "wallet-iso-2")
	suite.Require().NoError(err)
	suite.Require().Len(devices2, 1)
}

// =============================================================================
// RevokeDevice Extended Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestRevokeDevice_SetsRevokedAt() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-revoke-at", "device-ra", "Device", "mobile", []byte("fp-ra"))
	suite.Require().NoError(err)

	// Revoke
	err = k.RevokeDevice(ctx, "wallet-revoke-at", "device-ra")
	suite.Require().NoError(err)

	// Get device and verify RevokedAt is set
	devices, err := k.GetDevices(ctx, "wallet-revoke-at")
	suite.Require().NoError(err)
	suite.Require().Len(devices, 1)
	suite.Require().False(devices[0].Trusted)
	suite.Require().NotNil(devices[0].RevokedAt)
	suite.Require().Greater(devices[0].RevokedAt.Seconds, int64(0))
}

func (suite *DeviceFingerprintTestSuite) TestRevokeDevice_DoubleRevoke() {
	ctx := suite.GetContext()
	k := suite.GetKeeper()

	// Register device
	_, err := k.RegisterDevice(ctx, "wallet-double-revoke", "device-dr", "Device", "mobile", []byte("fp-dr"))
	suite.Require().NoError(err)

	// First revoke
	err = k.RevokeDevice(ctx, "wallet-double-revoke", "device-dr")
	suite.Require().NoError(err)

	// Second revoke should still work (idempotent)
	err = k.RevokeDevice(ctx, "wallet-double-revoke", "device-dr")
	suite.Require().NoError(err)

	// Device should still be revoked
	devices, err := k.GetDevices(ctx, "wallet-double-revoke")
	suite.Require().NoError(err)
	suite.Require().Len(devices, 1)
	suite.Require().False(devices[0].Trusted)
}

// =============================================================================
// Hash Function Tests
// =============================================================================

func (suite *DeviceFingerprintTestSuite) TestHashFingerprint_Empty() {
	k := suite.GetKeeper()

	hash := k.hashFingerprint([]byte{})
	suite.Require().NotEmpty(hash)
	suite.Require().Len(hash, 32)
}

func (suite *DeviceFingerprintTestSuite) TestHashFingerprint_Deterministic() {
	k := suite.GetKeeper()

	fingerprint := []byte("deterministic-test")

	// Hash same input 100 times
	expectedHash := k.hashFingerprint(fingerprint)
	for i := 0; i < 100; i++ {
		hash := k.hashFingerprint(fingerprint)
		suite.Require().Equal(expectedHash, hash, "hash should be deterministic")
	}
}

func (suite *DeviceFingerprintTestSuite) TestHashFingerprint_CollisionResistance() {
	k := suite.GetKeeper()

	// Test that similar inputs produce different hashes
	inputs := [][]byte{
		[]byte("test"),
		[]byte("Test"),
		[]byte("test "),
		[]byte(" test"),
		[]byte("test1"),
		[]byte("tEst"),
	}

	hashes := make(map[string]bool)
	for _, input := range inputs {
		hash := k.hashFingerprint(input)
		hashStr := string(hash)
		suite.Require().False(hashes[hashStr], "hash collision detected for input: %s", input)
		hashes[hashStr] = true
	}
}
