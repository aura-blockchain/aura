// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

func TestQueryServer(t *testing.T) {
	k, ctx := setupKeeper(t)
	queryServer := keeper.NewQueryServerImpl(&k)

	t.Run("Params - success", func(t *testing.T) {
		req := &cryptoproto.QueryParamsRequest{}
		resp, err := queryServer.Params(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Params)
	})

	t.Run("Params - nil request", func(t *testing.T) {
		_, err := queryServer.Params(ctx, nil)
		require.Error(t, err)
	})

	t.Run("KeyRotationSchedule - success", func(t *testing.T) {
		scheduleID, err := k.CreateKeyRotationSchedule(ctx, "creator", "test-key", 86400, nil)
		require.NoError(t, err)

		req := &cryptoproto.QueryKeyRotationScheduleRequest{
			Id: scheduleID,
		}
		resp, err := queryServer.KeyRotationSchedule(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Schedule)
		require.Equal(t, scheduleID, resp.Schedule.Id)
	})

	t.Run("KeyRotationSchedule - nil request", func(t *testing.T) {
		_, err := queryServer.KeyRotationSchedule(ctx, nil)
		require.Error(t, err)
	})

	t.Run("KeyRotationSchedule - not found", func(t *testing.T) {
		req := &cryptoproto.QueryKeyRotationScheduleRequest{
			Id: "non-existent",
		}
		_, err := queryServer.KeyRotationSchedule(ctx, req)
		require.Error(t, err)
	})

	t.Run("ThresholdScheme - success", func(t *testing.T) {
		schemeID, _, err := k.CreateThresholdScheme(ctx, "creator", 2, 3, []string{"p1", "p2", "p3"}, cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA)
		require.NoError(t, err)

		req := &cryptoproto.QueryThresholdSchemeRequest{
			SchemeId: schemeID,
		}
		resp, err := queryServer.ThresholdScheme(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Scheme)
		require.Equal(t, schemeID, resp.Scheme.SchemeId)
	})

	t.Run("ThresholdScheme - nil request", func(t *testing.T) {
		_, err := queryServer.ThresholdScheme(ctx, nil)
		require.Error(t, err)
	})

	t.Run("ThresholdScheme - not found", func(t *testing.T) {
		req := &cryptoproto.QueryThresholdSchemeRequest{
			SchemeId: "non-existent",
		}
		_, err := queryServer.ThresholdScheme(ctx, req)
		require.Error(t, err)
	})

	t.Run("VerifyZKProof - nil request", func(t *testing.T) {
		_, err := queryServer.VerifyZKProof(ctx, nil)
		require.Error(t, err)
	})

	t.Run("VerifyZKProof - not implemented", func(t *testing.T) {
		req := &cryptoproto.QueryVerifyZKProofRequest{
			ProofId:      "proof-1",
			ProofData:    make([]byte, 32),
			PublicInputs: make([]byte, 32),
		}
		resp, err := queryServer.VerifyZKProof(ctx, req)
		require.NoError(t, err) // Returns error in response, not as error
		require.False(t, resp.Valid)
		require.NotEmpty(t, resp.ErrorMessage)
	})

	t.Run("SecureEnclave - success", func(t *testing.T) {
		attestation := make([]byte, 432)
		enclaveID, err := k.RegisterSecureEnclave(ctx, "creator", cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX, attestation, nil)
		require.NoError(t, err)

		req := &cryptoproto.QuerySecureEnclaveRequest{
			EnclaveId: enclaveID,
		}
		resp, err := queryServer.SecureEnclave(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Enclave)
		require.Equal(t, enclaveID, resp.Enclave.EnclaveId)
	})

	t.Run("SecureEnclave - nil request", func(t *testing.T) {
		_, err := queryServer.SecureEnclave(ctx, nil)
		require.Error(t, err)
	})

	t.Run("SecureEnclave - not found", func(t *testing.T) {
		req := &cryptoproto.QuerySecureEnclaveRequest{
			EnclaveId: "non-existent",
		}
		_, err := queryServer.SecureEnclave(ctx, req)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrSecureEnclaveNotFound)
	})

	t.Run("QuantumResistantKey - success", func(t *testing.T) {
		publicKey := GenerateDummyQuantumPublicKey(cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM)
		keyID, err := k.RegisterQuantumResistantKey(ctx, "creator", cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM, publicKey, nil)
		require.NoError(t, err)

		req := &cryptoproto.QueryQuantumResistantKeyRequest{
			KeyId: keyID,
		}
		resp, err := queryServer.QuantumResistantKey(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Key)
		require.Equal(t, keyID, resp.Key.KeyId)
	})

	t.Run("QuantumResistantKey - nil request", func(t *testing.T) {
		_, err := queryServer.QuantumResistantKey(ctx, nil)
		require.Error(t, err)
	})

	t.Run("QuantumResistantKey - not found", func(t *testing.T) {
		req := &cryptoproto.QueryQuantumResistantKeyRequest{
			KeyId: "non-existent",
		}
		_, err := queryServer.QuantumResistantKey(ctx, req)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrQuantumKeyNotFound)
	})

	t.Run("RandomSourceStatus - success", func(t *testing.T) {
		entropy := make([]byte, 64)
		_, err := k.InitializeRandomSource(ctx, cryptoproto.RandomSourceType_RANDOM_SOURCE_TYPE_SYSTEM, entropy)
		require.NoError(t, err)

		req := &cryptoproto.QueryRandomSourceStatusRequest{}
		resp, err := queryServer.RandomSourceStatus(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Sources)
	})

	t.Run("RandomSourceStatus - nil request", func(t *testing.T) {
		_, err := queryServer.RandomSourceStatus(ctx, nil)
		require.Error(t, err)
	})

	t.Run("CertificatePin - success", func(t *testing.T) {
		hash := make([]byte, 32)
		_, err := k.AddCertificatePin(ctx, "creator", "query-test.com", [][]byte{hash}, cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI, nil)
		require.NoError(t, err)

		req := &cryptoproto.QueryCertificatePinRequest{
			Hostname: "query-test.com",
		}
		resp, err := queryServer.CertificatePin(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Pin)
		require.Equal(t, "query-test.com", resp.Pin.Hostname)
	})

	t.Run("CertificatePin - nil request", func(t *testing.T) {
		_, err := queryServer.CertificatePin(ctx, nil)
		require.Error(t, err)
	})

	t.Run("CertificatePin - not found", func(t *testing.T) {
		req := &cryptoproto.QueryCertificatePinRequest{
			Hostname: "non-existent.com",
		}
		_, err := queryServer.CertificatePin(ctx, req)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrCertificatePinNotFound)
	})
}
