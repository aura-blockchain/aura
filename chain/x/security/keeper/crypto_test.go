// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/testutil"
	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

// =============================================================================
// Key Rotation Schedule Tests
// =============================================================================

func TestDeleteKeyRotationSchedule(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "test-schedule",
		KeyId:                   "key-1",
		RotationIntervalSeconds: 86400,
		Enabled:                 true,
	}

	// Set the schedule
	k.SetKeyRotationSchedule(ctx, schedule)

	// Verify it exists
	retrieved, found := k.GetKeyRotationSchedule(ctx, "test-schedule")
	require.True(t, found)
	require.Equal(t, "test-schedule", retrieved.Id)

	// Delete it
	k.DeleteKeyRotationSchedule(ctx, "test-schedule")

	// Verify it's gone
	_, found = k.GetKeyRotationSchedule(ctx, "test-schedule")
	require.False(t, found)
}

func TestDeleteKeyRotationScheduleNonExistent(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Deleting a non-existent schedule should not panic
	k.DeleteKeyRotationSchedule(ctx, "non-existent")

	// Verify no schedule exists
	_, found := k.GetKeyRotationSchedule(ctx, "non-existent")
	require.False(t, found)
}

// =============================================================================
// Secure Enclave Tests
// =============================================================================

func TestSetGetSecureEnclave(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	enclave := &types.SecureEnclave{
		Id:           "enclave-1",
		EnclaveType:  "SGX",
		PublicKey:    "0x1234567890abcdef",
		Attestation:  "attestation-data",
		RegisteredAt: &now,
		IsVerified:   true,
	}

	// Set the enclave
	k.SetSecureEnclave(ctx, enclave)

	// Get and verify
	retrieved, found := k.GetSecureEnclave(ctx, "enclave-1")
	require.True(t, found)
	require.Equal(t, "enclave-1", retrieved.Id)
	require.Equal(t, "SGX", retrieved.EnclaveType)
	require.Equal(t, "0x1234567890abcdef", retrieved.PublicKey)
	require.True(t, retrieved.IsVerified)
}

func TestGetSecureEnclaveNotFound(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Get non-existent enclave
	_, found := k.GetSecureEnclave(ctx, "non-existent")
	require.False(t, found)
}

func TestGetAllSecureEnclaves(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Initially empty
	enclaves := k.GetAllSecureEnclaves(ctx)
	require.Empty(t, enclaves)

	// Add multiple enclaves
	now := time.Now()
	for i := 1; i <= 3; i++ {
		enclave := &types.SecureEnclave{
			Id:           "enclave-" + string(rune('0'+i)),
			EnclaveType:  "SGX",
			RegisteredAt: &now,
		}
		k.SetSecureEnclave(ctx, enclave)
	}

	// Get all
	enclaves = k.GetAllSecureEnclaves(ctx)
	require.Len(t, enclaves, 3)
}

// =============================================================================
// Quantum Resistant Key Tests
// =============================================================================

func TestGetQuantumResistantKey(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	qrk := &securitypb.QuantumResistantKey{
		KeyId:       "qrk-1",
		Algorithm:   securitypb.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER,
		PublicKey:   []byte("quantum-public-key"),
		KeyMetadata: []byte("metadata"),
		CreatedAt:   now,
	}

	// Set the key
	k.SetQuantumResistantKey(ctx, qrk)

	// Get and verify
	retrieved, found := k.GetQuantumResistantKey(ctx, "qrk-1")
	require.True(t, found)
	require.Equal(t, "qrk-1", retrieved.KeyId)
	require.Equal(t, securitypb.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER, retrieved.Algorithm)
	require.Equal(t, []byte("quantum-public-key"), retrieved.PublicKey)
}

func TestGetQuantumResistantKeyNotFound(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	_, found := k.GetQuantumResistantKey(ctx, "non-existent")
	require.False(t, found)
}

// =============================================================================
// Random Source Tests
// =============================================================================

func TestSetGetRandomSource(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	source := &types.RandomSource{
		Id:           "rand-1",
		SourceType:   "VRF",
		Endpoint:     "https://randomness.example.com",
		PublicKey:    "vrf-public-key",
		LastVerified: &now,
		IsActive:     true,
	}

	// Set the source
	k.SetRandomSource(ctx, source)

	// Get and verify
	retrieved, found := k.GetRandomSource(ctx, "rand-1")
	require.True(t, found)
	require.Equal(t, "rand-1", retrieved.Id)
	require.Equal(t, "VRF", retrieved.SourceType)
	require.True(t, retrieved.IsActive)
}

func TestGetRandomSourceNotFound(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	_, found := k.GetRandomSource(ctx, "non-existent")
	require.False(t, found)
}

func TestGetAllRandomSources(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Initially empty
	sources := k.GetAllRandomSources(ctx)
	require.Empty(t, sources)

	// Add multiple sources
	now := time.Now()
	for i := 1; i <= 5; i++ {
		source := &types.RandomSource{
			Id:           "rand-" + string(rune('0'+i)),
			SourceType:   "VRF",
			LastVerified: &now,
		}
		k.SetRandomSource(ctx, source)
	}

	// Get all
	sources = k.GetAllRandomSources(ctx)
	require.Len(t, sources, 5)
}

// =============================================================================
// Certificate Pin Tests
// =============================================================================

func TestSetGetCertificatePin(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	now := time.Now()
	validUntil := now.Add(365 * 24 * time.Hour)
	pin := &types.CertificatePin{
		Id:         "pin-1",
		Domain:     "api.example.com",
		PinHash:    "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		PinType:    "SPKI",
		ValidFrom:  &now,
		ValidUntil: &validUntil,
		IsActive:   true,
	}

	// Set the pin
	k.SetCertificatePin(ctx, pin)

	// Get and verify
	retrieved, found := k.GetCertificatePin(ctx, "pin-1")
	require.True(t, found)
	require.Equal(t, "pin-1", retrieved.Id)
	require.Equal(t, "api.example.com", retrieved.Domain)
	require.True(t, retrieved.IsActive)
}

func TestGetCertificatePinNotFound(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	_, found := k.GetCertificatePin(ctx, "non-existent")
	require.False(t, found)
}

func TestGetAllCertificatePins(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Initially empty
	pins := k.GetAllCertificatePins(ctx)
	require.Empty(t, pins)

	// Add multiple pins
	for i := 1; i <= 4; i++ {
		pin := &types.CertificatePin{
			Id:       "pin-" + string(rune('0'+i)),
			Domain:   "domain-" + string(rune('0'+i)) + ".example.com",
			IsActive: true,
		}
		k.SetCertificatePin(ctx, pin)
	}

	// Get all
	pins = k.GetAllCertificatePins(ctx)
	require.Len(t, pins, 4)
}

func TestDeleteCertificatePin(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	pin := &types.CertificatePin{
		Id:       "pin-to-delete",
		Domain:   "delete.example.com",
		IsActive: true,
	}

	// Set the pin
	k.SetCertificatePin(ctx, pin)

	// Verify it exists
	_, found := k.GetCertificatePin(ctx, "pin-to-delete")
	require.True(t, found)

	// Delete it
	k.DeleteCertificatePin(ctx, "pin-to-delete")

	// Verify it's gone
	_, found = k.GetCertificatePin(ctx, "pin-to-delete")
	require.False(t, found)
}

// =============================================================================
// Check Key Rotation Due Tests
// =============================================================================

func TestCheckKeyRotationDue(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	// Create schedules: some due, some not
	pastTime := baseTime.Add(-1 * time.Hour)
	futureTime := baseTime.Add(1 * time.Hour)

	dueSchedule := &securitypb.KeyRotationSchedule{
		Id:                      "due-schedule",
		KeyId:                   "key-1",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        pastTime,
		CreatedBy:               "admin",
		Enabled:                 true,
	}

	notDueSchedule := &securitypb.KeyRotationSchedule{
		Id:                      "not-due-schedule",
		KeyId:                   "key-2",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        futureTime,
		CreatedBy:               "admin",
		Enabled:                 true,
	}

	disabledSchedule := &securitypb.KeyRotationSchedule{
		Id:                      "disabled-schedule",
		KeyId:                   "key-3",
		RotationIntervalSeconds: 3600,
		NextRotationTime:        pastTime,
		CreatedBy:               "admin",
		Enabled:                 false,
	}

	k.SetKeyRotationSchedule(ctx, dueSchedule)
	k.SetKeyRotationSchedule(ctx, notDueSchedule)
	k.SetKeyRotationSchedule(ctx, disabledSchedule)

	// Check which are due
	dueForRotation := k.CheckKeyRotationDue(ctx)

	// Only the enabled, past-due schedule should be returned
	require.Len(t, dueForRotation, 1)
	require.Equal(t, "due-schedule", dueForRotation[0].Id)
}

func TestCheckKeyRotationDueMaxLimit(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	pastTime := baseTime.Add(-1 * time.Hour)

	// Create more than MaxKeyRotationsPerBlock (50) schedules
	for i := 0; i < 60; i++ {
		id := "schedule-" + string(rune('A'+i/26)) + string(rune('A'+i%26))
		schedule := &securitypb.KeyRotationSchedule{
			Id:                      id,
			KeyId:                   "key-" + id,
			RotationIntervalSeconds: 3600,
			NextRotationTime:        pastTime,
			CreatedBy:               "admin",
			Enabled:                 true,
		}
		k.SetKeyRotationSchedule(ctx, schedule)
	}

	// Should only return up to 50
	dueForRotation := k.CheckKeyRotationDue(ctx)
	require.LessOrEqual(t, len(dueForRotation), 50)
}

func TestCheckKeyRotationDueEmpty(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// No schedules
	dueForRotation := k.CheckKeyRotationDue(ctx)
	require.Empty(t, dueForRotation)
}

// =============================================================================
// Validate Certificate Pin Tests
// =============================================================================

func TestValidateCertificatePinValid(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	validFrom := baseTime.Add(-1 * time.Hour)
	validUntil := baseTime.Add(1 * time.Hour)

	pin := &types.CertificatePin{
		Id:         "pin-valid",
		Domain:     "test.example.com",
		PinHash:    "sha256/validhash123",
		ValidFrom:  &validFrom,
		ValidUntil: &validUntil,
		IsActive:   true,
	}
	k.SetCertificatePin(ctx, pin)

	// Valid hash should pass
	err := k.ValidateCertificatePin(ctx, "test.example.com", "sha256/validhash123")
	require.NoError(t, err)
}

func TestValidateCertificatePinInvalidHash(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	validFrom := baseTime.Add(-1 * time.Hour)
	validUntil := baseTime.Add(1 * time.Hour)

	pin := &types.CertificatePin{
		Id:         "pin-hash-check",
		Domain:     "hash.example.com",
		PinHash:    "sha256/correcthash",
		ValidFrom:  &validFrom,
		ValidUntil: &validUntil,
		IsActive:   true,
	}
	k.SetCertificatePin(ctx, pin)

	// Wrong hash should fail
	err := k.ValidateCertificatePin(ctx, "hash.example.com", "sha256/wronghash")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrCertificateInvalid)
}

func TestValidateCertificatePinExpired(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	validFrom := baseTime.Add(-2 * time.Hour)
	validUntil := baseTime.Add(-1 * time.Hour) // Expired

	pin := &types.CertificatePin{
		Id:         "pin-expired",
		Domain:     "expired.example.com",
		PinHash:    "sha256/somehash",
		ValidFrom:  &validFrom,
		ValidUntil: &validUntil,
		IsActive:   true,
	}
	k.SetCertificatePin(ctx, pin)

	// Expired pin should fail
	err := k.ValidateCertificatePin(ctx, "expired.example.com", "sha256/somehash")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrCertificateInvalid)
}

func TestValidateCertificatePinNotYetValid(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// Set a specific block time for this test
	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(baseTime)

	validFrom := baseTime.Add(1 * time.Hour) // Not yet valid
	validUntil := baseTime.Add(2 * time.Hour)

	pin := &types.CertificatePin{
		Id:         "pin-future",
		Domain:     "future.example.com",
		PinHash:    "sha256/futurehash",
		ValidFrom:  &validFrom,
		ValidUntil: &validUntil,
		IsActive:   true,
	}
	k.SetCertificatePin(ctx, pin)

	// Not yet valid pin should fail
	err := k.ValidateCertificatePin(ctx, "future.example.com", "sha256/futurehash")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrCertificateInvalid)
}

func TestValidateCertificatePinNoPinForDomain(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	// No pin configured for this domain - should allow by default
	err := k.ValidateCertificatePin(ctx, "unpinned.example.com", "sha256/anyhash")
	require.NoError(t, err)
}

func TestValidateCertificatePinInactiveDomain(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	pin := &types.CertificatePin{
		Id:       "pin-inactive",
		Domain:   "inactive.example.com",
		PinHash:  "sha256/inactivehash",
		IsActive: false, // Inactive
	}
	k.SetCertificatePin(ctx, pin)

	// Inactive pin should be ignored - allow by default
	err := k.ValidateCertificatePin(ctx, "inactive.example.com", "sha256/wronghash")
	require.NoError(t, err)
}

// =============================================================================
// Key Rotation Execution Tests
// =============================================================================

func TestRotateKey(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	pastRotation := baseTime.Add(-24 * time.Hour)

	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "rotation-test",
		KeyId:                   "key-1",
		RotationIntervalSeconds: 86400, // 24 hours
		NextRotationTime:        baseTime,
		LastRotation:            &pastRotation,
		CreatedBy:               "admin",
		Enabled:                 true,
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	// Rotate the key
	err := k.RotateKey(ctx, "rotation-test")
	require.NoError(t, err)

	// Verify the schedule was updated
	updated, found := k.GetKeyRotationSchedule(ctx, "rotation-test")
	require.True(t, found)
	require.NotNil(t, updated.LastRotation)
}

func TestRotateKeyNotFound(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	err := k.RotateKey(ctx, "non-existent")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrKeyRotationNotFound)
}

func TestRotateKeyDisabled(t *testing.T) {
	ctx, k := testutil.NewSecurityKeeperForTest(t)

	baseTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	schedule := &securitypb.KeyRotationSchedule{
		Id:                      "disabled-rotation",
		KeyId:                   "key-disabled",
		RotationIntervalSeconds: 86400,
		NextRotationTime:        baseTime,
		CreatedBy:               "admin",
		Enabled:                 false, // Disabled
	}
	k.SetKeyRotationSchedule(ctx, schedule)

	err := k.RotateKey(ctx, "disabled-rotation")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrKeyRotationDisabled)
}
