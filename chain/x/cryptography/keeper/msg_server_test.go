package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/cryptography/keeper"
	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

func TestMsgServer(t *testing.T) {
	k, ctx := setupKeeper(t)
	msgServer := keeper.NewMsgServerImpl(&k)

	t.Run("CreateKeyRotationSchedule - success", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateKeyRotationSchedule{
			Creator:                 "creator",
			KeyId:                   "test-key-1",
			RotationIntervalSeconds: 86400,
			Policy:                  nil,
		}

		resp, err := msgServer.CreateKeyRotationSchedule(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.ScheduleId)
	})

	t.Run("CreateKeyRotationSchedule - empty key_id", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateKeyRotationSchedule{
			Creator:                 "creator",
			KeyId:                   "",
			RotationIntervalSeconds: 86400,
		}

		_, err := msgServer.CreateKeyRotationSchedule(ctx, msg)
		require.Error(t, err)
	})

	t.Run("CreateKeyRotationSchedule - invalid interval", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateKeyRotationSchedule{
			Creator:                 "creator",
			KeyId:                   "test-key",
			RotationIntervalSeconds: 0,
		}

		_, err := msgServer.CreateKeyRotationSchedule(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RotateKey - success", func(t *testing.T) {
		publicKey := make([]byte, 32)
		msg := &cryptoproto.MsgRotateKey{
			Creator:      "creator",
			KeyId:        "test-key-2",
			NewPublicKey: publicKey,
		}

		resp, err := msgServer.RotateKey(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.RotationId)
		require.NotNil(t, resp.RotationTime)
	})

	t.Run("RotateKey - empty key_id", func(t *testing.T) {
		msg := &cryptoproto.MsgRotateKey{
			Creator:      "creator",
			KeyId:        "",
			NewPublicKey: []byte("key"),
		}

		_, err := msgServer.RotateKey(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RotateKey - empty public key", func(t *testing.T) {
		msg := &cryptoproto.MsgRotateKey{
			Creator:      "creator",
			KeyId:        "test-key",
			NewPublicKey: []byte{},
		}

		_, err := msgServer.RotateKey(ctx, msg)
		require.Error(t, err)
	})

	t.Run("CreateThresholdScheme - success", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateThresholdScheme{
			Creator:           "creator",
			Threshold:         2,
			TotalParticipants: 3,
			ParticipantIds:    []string{"p1", "p2", "p3"},
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		}

		resp, err := msgServer.CreateThresholdScheme(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.SchemeId)
		require.NotEmpty(t, resp.PublicKey)
	})

	t.Run("CreateThresholdScheme - invalid threshold", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateThresholdScheme{
			Creator:           "creator",
			Threshold:         0,
			TotalParticipants: 3,
			ParticipantIds:    []string{"p1", "p2", "p3"},
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		}

		_, err := msgServer.CreateThresholdScheme(ctx, msg)
		require.Error(t, err)
	})

	t.Run("CreateThresholdScheme - threshold > participants", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateThresholdScheme{
			Creator:           "creator",
			Threshold:         5,
			TotalParticipants: 3,
			ParticipantIds:    []string{"p1", "p2", "p3"},
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		}

		_, err := msgServer.CreateThresholdScheme(ctx, msg)
		require.Error(t, err)
	})

	t.Run("CreateThresholdScheme - mismatched participant count", func(t *testing.T) {
		msg := &cryptoproto.MsgCreateThresholdScheme{
			Creator:           "creator",
			Threshold:         2,
			TotalParticipants: 5,
			ParticipantIds:    []string{"p1", "p2", "p3"},
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		}

		_, err := msgServer.CreateThresholdScheme(ctx, msg)
		require.Error(t, err)
	})

	t.Run("SubmitThresholdSignatureShare - success", func(t *testing.T) {
		// First create a scheme
		createMsg := &cryptoproto.MsgCreateThresholdScheme{
			Creator:           "creator",
			Threshold:         2,
			TotalParticipants: 3,
			ParticipantIds:    []string{"p1", "p2", "p3"},
			SchemeType:        cryptoproto.ThresholdSchemeType_THRESHOLD_SCHEME_TYPE_ECDSA,
		}
		createResp, err := msgServer.CreateThresholdScheme(ctx, createMsg)
		require.NoError(t, err)

		// Submit share
		msg := &cryptoproto.MsgSubmitThresholdSignatureShare{
			Submitter:      "p1",
			SchemeId:       createResp.SchemeId,
			SignatureShare: make([]byte, 64),
			MessageHash:    make([]byte, 32),
		}

		resp, err := msgServer.SubmitThresholdSignatureShare(ctx, msg)
		require.NoError(t, err)
		require.Equal(t, int32(1), resp.SharesCollected)
		require.False(t, resp.ThresholdReached)
	})

	t.Run("SubmitThresholdSignatureShare - empty scheme_id", func(t *testing.T) {
		msg := &cryptoproto.MsgSubmitThresholdSignatureShare{
			Submitter:      "p1",
			SchemeId:       "",
			SignatureShare: make([]byte, 64),
			MessageHash:    make([]byte, 32),
		}

		_, err := msgServer.SubmitThresholdSignatureShare(ctx, msg)
		require.Error(t, err)
	})

	t.Run("SubmitThresholdSignatureShare - empty signature", func(t *testing.T) {
		msg := &cryptoproto.MsgSubmitThresholdSignatureShare{
			Submitter:      "p1",
			SchemeId:       "scheme-1",
			SignatureShare: []byte{},
			MessageHash:    make([]byte, 32),
		}

		_, err := msgServer.SubmitThresholdSignatureShare(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RegisterZKProofCircuit - not implemented", func(t *testing.T) {
		msg := &cryptoproto.MsgRegisterZKProofCircuit{
			Creator:          "creator",
			CircuitId:        "circuit-1",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
			PublicParameters: make([]byte, 32),
			VerificationKey:  make([]byte, 32),
		}

		_, err := msgServer.RegisterZKProofCircuit(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RegisterZKProofCircuit - empty circuit_id", func(t *testing.T) {
		msg := &cryptoproto.MsgRegisterZKProofCircuit{
			Creator:          "creator",
			CircuitId:        "",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
			PublicParameters: make([]byte, 32),
			VerificationKey:  make([]byte, 32),
		}

		_, err := msgServer.RegisterZKProofCircuit(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RegisterZKProofCircuit - empty verification key", func(t *testing.T) {
		msg := &cryptoproto.MsgRegisterZKProofCircuit{
			Creator:          "creator",
			CircuitId:        "circuit-1",
			ProofType:        cryptoproto.ZKProofType_ZK_PROOF_TYPE_GROTH16,
			PublicParameters: make([]byte, 32),
			VerificationKey:  []byte{},
		}

		_, err := msgServer.RegisterZKProofCircuit(ctx, msg)
		require.Error(t, err)
	})

	t.Run("SubmitZKProof - not implemented", func(t *testing.T) {
		msg := &cryptoproto.MsgSubmitZKProof{
			Submitter:    "submitter",
			ProofId:      "proof-1",
			ProofData:    make([]byte, 32),
			PublicInputs: make([]byte, 32),
		}

		_, err := msgServer.SubmitZKProof(ctx, msg)
		require.Error(t, err)
	})

	t.Run("SubmitZKProof - empty proof_id", func(t *testing.T) {
		msg := &cryptoproto.MsgSubmitZKProof{
			Submitter:    "submitter",
			ProofId:      "",
			ProofData:    make([]byte, 32),
			PublicInputs: make([]byte, 32),
		}

		_, err := msgServer.SubmitZKProof(ctx, msg)
		require.Error(t, err)
	})

	t.Run("SubmitZKProof - empty proof data", func(t *testing.T) {
		msg := &cryptoproto.MsgSubmitZKProof{
			Submitter:    "submitter",
			ProofId:      "proof-1",
			ProofData:    []byte{},
			PublicInputs: make([]byte, 32),
		}

		_, err := msgServer.SubmitZKProof(ctx, msg)
		require.Error(t, err)
	})

	t.Run("RegisterSecureEnclave - success", func(t *testing.T) {
		msg := &cryptoproto.MsgRegisterSecureEnclave{
			Creator:         "creator",
			EnclaveType:     cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			AttestationData: make([]byte, 432),
			EnclaveMetadata: map[string]string{"version": "1.0"},
		}

		resp, err := msgServer.RegisterSecureEnclave(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.EnclaveId)
	})

	t.Run("RegisterSecureEnclave - empty attestation", func(t *testing.T) {
		msg := &cryptoproto.MsgRegisterSecureEnclave{
			Creator:         "creator",
			EnclaveType:     cryptoproto.SecureEnclaveType_SECURE_ENCLAVE_TYPE_SGX,
			AttestationData: []byte{},
			EnclaveMetadata: nil,
		}

		_, err := msgServer.RegisterSecureEnclave(ctx, msg)
		require.Error(t, err)
	})

	t.Run("GenerateQuantumResistantKey - success", func(t *testing.T) {
		expiresAt := time.Now().Add(365 * 24 * time.Hour)
		msg := &cryptoproto.MsgGenerateQuantumResistantKey{
			Creator:   "creator",
			Algorithm: cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM,
			ExpiresAt: timestamppb.New(expiresAt),
		}

		resp, err := msgServer.GenerateQuantumResistantKey(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.KeyId)
		require.NotEmpty(t, resp.PublicKey)
	})

	t.Run("GenerateQuantumResistantKey - nil expires_at", func(t *testing.T) {
		msg := &cryptoproto.MsgGenerateQuantumResistantKey{
			Creator:   "creator",
			Algorithm: cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER,
			ExpiresAt: nil,
		}

		resp, err := msgServer.GenerateQuantumResistantKey(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.KeyId)
	})

	t.Run("AddCertificatePin - success", func(t *testing.T) {
		hash := make([]byte, 32)
		expiresAt := time.Now().Add(365 * 24 * time.Hour)
		msg := &cryptoproto.MsgAddCertificatePin{
			Creator:           "creator",
			Hostname:          "test.example.com",
			CertificateHashes: [][]byte{hash},
			PinType:           cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			ExpiresAt:         timestamppb.New(expiresAt),
		}

		resp, err := msgServer.AddCertificatePin(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.PinId)
	})

	t.Run("AddCertificatePin - empty hostname", func(t *testing.T) {
		hash := make([]byte, 32)
		msg := &cryptoproto.MsgAddCertificatePin{
			Creator:           "creator",
			Hostname:          "",
			CertificateHashes: [][]byte{hash},
			PinType:           cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
		}

		_, err := msgServer.AddCertificatePin(ctx, msg)
		require.Error(t, err)
	})

	t.Run("AddCertificatePin - empty hashes", func(t *testing.T) {
		msg := &cryptoproto.MsgAddCertificatePin{
			Creator:           "creator",
			Hostname:          "test.com",
			CertificateHashes: [][]byte{},
			PinType:           cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
		}

		_, err := msgServer.AddCertificatePin(ctx, msg)
		require.Error(t, err)
	})

	t.Run("AddCertificatePin - nil expires_at", func(t *testing.T) {
		hash := make([]byte, 32)
		msg := &cryptoproto.MsgAddCertificatePin{
			Creator:           "creator",
			Hostname:          "test2.example.com",
			CertificateHashes: [][]byte{hash},
			PinType:           cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI,
			ExpiresAt:         nil,
		}

		resp, err := msgServer.AddCertificatePin(ctx, msg)
		require.NoError(t, err)
		require.NotEmpty(t, resp.PinId)
	})

	t.Run("UpdateParams - success", func(t *testing.T) {
		newParams := types.DefaultParams()
		newParams.DefaultRotationIntervalDays = 120

		msg := &cryptoproto.MsgUpdateParams{
			Authority: "authority",
			Params:    &newParams,
		}

		resp, err := msgServer.UpdateParams(ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("UpdateParams - unauthorized", func(t *testing.T) {
		newParams := types.DefaultParams()
		msg := &cryptoproto.MsgUpdateParams{
			Authority: "wrong-authority",
			Params:    &newParams,
		}

		_, err := msgServer.UpdateParams(ctx, msg)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrUnauthorized)
	})
}
