// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// RegisterSecureEnclave registers a secure enclave for key storage
func (k Keeper) RegisterSecureEnclave(
	ctx context.Context,
	creator string,
	enclaveType cryptoproto.SecureEnclaveType,
	attestationData []byte,
	enclaveMetadata map[string]string,
) (string, error) {
	if enclaveType == cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_UNSPECIFIED {
		return "", fmt.Errorf("enclave type must be specified")
	}
	if len(attestationData) == 0 {
		return "", fmt.Errorf("attestation data required")
	}

	// Verify attestation
	if err := k.verifyEnclaveAttestation(enclaveType, attestationData); err != nil {
		return "", types.ErrEnclaveAttestationFailed
	}

	// Generate enclave ID
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	enclaveID := fmt.Sprintf("enclave_%s_%d", enclaveType.String(), blockTime.Unix())

	enclave := &cryptoproto.SecureEnclaveConfig{
		EnclaveId:       enclaveID,
		EnclaveType:     enclaveType,
		AttestationData: attestationData,
		AttestationTime: blockTime,
		Status:          cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY,
		EnclaveMetadata: enclaveMetadata,
	}

	// Store in KV store
	if err := k.SetSecureEnclaveConfig(ctx, enclave); err != nil {
		return "", err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("registered secure enclave",
		"enclave_id", enclaveID,
		"type", enclaveType.String(),
	)

	return enclaveID, nil
}

// Note: GetSecureEnclave is now implemented in keeper.go using KV store

// verifyEnclaveAttestation verifies secure enclave attestation data
func (k Keeper) verifyEnclaveAttestation(
	enclaveType cryptoproto.SecureEnclaveType,
	attestationData []byte,
) error {
	// In a real implementation, this would perform proper attestation verification
	// based on the enclave type

	switch enclaveType {
	case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX:
		return k.verifySGXAttestation(attestationData)
	case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SEV:
		return k.verifySEVAttestation(attestationData)
	case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_TPM:
		return k.verifyTPMAttestation(attestationData)
	case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_HSM:
		return k.verifyHSMAttestation(attestationData)
	case cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_KEYCHAIN:
		return k.verifyKeychainAttestation(attestationData)
	default:
		return fmt.Errorf("unsupported enclave type")
	}
}

// verifySGXAttestation verifies Intel SGX attestation
func (k Keeper) verifySGXAttestation(attestationData []byte) error {
	// In a real implementation, verify SGX quote and report
	// Check MRENCLAVE, MRSIGNER, and other attestation fields
	if len(attestationData) < 432 { // SGX quote minimum size
		return types.ErrEnclaveAttestationFailed
	}
	return nil
}

// verifySEVAttestation verifies AMD SEV attestation
func (k Keeper) verifySEVAttestation(attestationData []byte) error {
	// In a real implementation, verify SEV attestation report
	if len(attestationData) < 144 { // SEV attestation report size
		return types.ErrEnclaveAttestationFailed
	}
	return nil
}

// verifyTPMAttestation verifies TPM attestation
func (k Keeper) verifyTPMAttestation(attestationData []byte) error {
	// In a real implementation, verify TPM quote and PCR values
	if len(attestationData) < 48 {
		return types.ErrEnclaveAttestationFailed
	}
	return nil
}

// verifyHSMAttestation verifies HSM attestation
func (k Keeper) verifyHSMAttestation(attestationData []byte) error {
	// In a real implementation, verify HSM certificate and attestation
	if len(attestationData) < 32 {
		return types.ErrEnclaveAttestationFailed
	}
	return nil
}

// verifyKeychainAttestation verifies system keychain attestation
func (k Keeper) verifyKeychainAttestation(attestationData []byte) error {
	// In a real implementation, verify keychain access token
	if len(attestationData) < 16 {
		return types.ErrEnclaveAttestationFailed
	}
	return nil
}

// SealDataToEnclave seals data to a secure enclave
func (k Keeper) SealDataToEnclave(
	ctx context.Context,
	enclaveID string,
	data []byte,
) ([]byte, error) {
	enclave, err := k.GetSecureEnclave(ctx, enclaveID)
	if err != nil {
		return nil, err
	}

	if enclave.Status != cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY {
		return nil, fmt.Errorf("enclave not ready")
	}

	// In a real implementation, this would use the enclave's sealing key
	// For now, use a simple encryption scheme
	sealedData := make([]byte, len(data)+32)
	copy(sealedData[:32], enclave.AttestationData[:32])
	for i := range data {
		sealedData[i+32] = data[i] ^ sealedData[i%32]
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("sealed data to enclave",
		"enclave_id", enclaveID,
		"data_length", len(data),
	)

	return sealedData, nil
}

// UnsealDataFromEnclave unseals data from a secure enclave
func (k Keeper) UnsealDataFromEnclave(
	ctx context.Context,
	enclaveID string,
	sealedData []byte,
) ([]byte, error) {
	enclave, err := k.GetSecureEnclave(ctx, enclaveID)
	if err != nil {
		return nil, err
	}

	if enclave.Status != cryptoproto.SecureEnclaveStatus_SECURE_ENCLAVE_STATUS_READY {
		return nil, fmt.Errorf("enclave not ready")
	}

	if len(sealedData) < 32 {
		return nil, fmt.Errorf("invalid sealed data")
	}

	// In a real implementation, this would use the enclave's unsealing key
	data := make([]byte, len(sealedData)-32)
	for i := range data {
		data[i] = sealedData[i+32] ^ enclave.AttestationData[i%32]
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("unsealed data from enclave",
		"enclave_id", enclaveID,
		"data_length", len(data),
	)

	return data, nil
}

// UpdateEnclaveStatus updates the status of a secure enclave
func (k Keeper) UpdateEnclaveStatus(
	ctx context.Context,
	enclaveID string,
	status cryptoproto.SecureEnclaveStatus,
) error {
	enclave, err := k.GetSecureEnclave(ctx, enclaveID)
	if err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	enclave.Status = status

	if err := k.SetSecureEnclaveConfig(ctx, enclave); err != nil {
		return fmt.Errorf("failed to get: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.Logger(sdkCtx).Info("updated enclave status",
		"enclave_id", enclaveID,
		"status", status.String(),
	)

	return nil
}

// Note: ListSecureEnclaves is now implemented in keeper.go using KV store

// RemoteAttestEnclave performs remote attestation for an enclave
func (k Keeper) RemoteAttestEnclave(
	ctx context.Context,
	enclaveID string,
) ([]byte, error) {
	enclave, err := k.GetSecureEnclave(ctx, enclaveID)
	if err != nil {
		return nil, err
	}

	// Generate attestation report
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime()
	h := sha256.New()
	h.Write(enclave.AttestationData)
	h.Write([]byte(enclaveID))
	h.Write([]byte(blockTime.Format(time.RFC3339)))
	report := h.Sum(nil)

	k.Logger(sdkCtx).Info("generated remote attestation",
		"enclave_id", enclaveID,
	)

	return report, nil
}

// Note: SetSecureEnclaveConfig is now implemented in keeper.go using KV store
