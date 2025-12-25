// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ cryptoproto.MsgServer = msgServer{}

type msgServer struct {
	cryptoproto.UnimplementedMsgServer
	Keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) cryptoproto.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (ms msgServer) CreateKeyRotationSchedule(goCtx context.Context, msg *cryptoproto.MsgCreateKeyRotationSchedule) (*cryptoproto.MsgCreateKeyRotationScheduleResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.KeyId == "" || msg.RotationIntervalSeconds <= 0 {
		return nil, types.ErrInvalidInput.Wrap("key_id and rotation_interval are required")
	}

	scheduleID, err := ms.Keeper.CreateKeyRotationSchedule(ctx, msg.Creator, msg.KeyId, msg.RotationIntervalSeconds, msg.Policy)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgCreateKeyRotationScheduleResponse{ScheduleId: scheduleID}, nil
}

func (ms msgServer) RotateKey(goCtx context.Context, msg *cryptoproto.MsgRotateKey) (*cryptoproto.MsgRotateKeyResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.KeyId == "" || len(msg.NewPublicKey) == 0 {
		return nil, types.ErrInvalidInput.Wrap("key_id and new_public_key are required")
	}

	// Verify ownership of key rotation schedules
	// Check if any schedules exist for this key and verify the creator owns them
	schedules := ms.Keeper.GetSchedulesForKey(ctx, msg.KeyId)
	for _, schedule := range schedules {
		if schedule.CreatedBy != msg.Creator {
			return nil, status.Error(codes.PermissionDenied, "not authorized to rotate this key - not the owner of associated rotation schedule")
		}
	}

	rotationID, rotationTime, err := ms.Keeper.RotateKey(ctx, msg.Creator, msg.KeyId, msg.NewPublicKey)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgRotateKeyResponse{
		RotationId:   rotationID,
		RotationTime: rotationTime,
	}, nil
}

func (ms msgServer) CreateThresholdScheme(goCtx context.Context, msg *cryptoproto.MsgCreateThresholdScheme) (*cryptoproto.MsgCreateThresholdSchemeResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Threshold <= 0 || msg.TotalParticipants <= 0 || msg.Threshold > msg.TotalParticipants {
		return nil, types.ErrInvalidInput.Wrap("invalid threshold or participant count")
	}

	if len(msg.ParticipantIds) != int(msg.TotalParticipants) {
		return nil, types.ErrInvalidInput.Wrap("participant_ids count must match total_participants")
	}

	schemeID, publicKey, err := ms.Keeper.CreateThresholdScheme(ctx, msg.Creator, uint32(msg.Threshold), uint32(msg.TotalParticipants), msg.ParticipantIds, msg.SchemeType)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgCreateThresholdSchemeResponse{
		SchemeId:  schemeID,
		PublicKey: publicKey,
	}, nil
}

func (ms msgServer) SubmitThresholdSignatureShare(goCtx context.Context, msg *cryptoproto.MsgSubmitThresholdSignatureShare) (*cryptoproto.MsgSubmitThresholdSignatureShareResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submitter address")
	}

	if !signers[0].Equals(submitterAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match submitter")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.SchemeId == "" || len(msg.SignatureShare) == 0 || len(msg.MessageHash) == 0 {
		return nil, types.ErrInvalidInput.Wrap("scheme_id, signature_share, and message_hash are required")
	}

	sharesCollected, thresholdReached, combinedSignature, err := ms.Keeper.SubmitThresholdSignatureShare(ctx, msg.Submitter, msg.SchemeId, msg.SignatureShare, msg.MessageHash)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgSubmitThresholdSignatureShareResponse{
		SharesCollected:   int32(sharesCollected),
		ThresholdReached:  thresholdReached,
		CombinedSignature: combinedSignature,
	}, nil
}

func (ms msgServer) RegisterZKProofCircuit(goCtx context.Context, msg *cryptoproto.MsgRegisterZKProofCircuit) (*cryptoproto.MsgRegisterZKProofCircuitResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.CircuitId == "" || len(msg.VerificationKey) == 0 {
		return nil, types.ErrInvalidInput.Wrap("circuit_id and verification_key are required")
	}

	proofID, err := ms.Keeper.RegisterZKProofCircuit(ctx, msg.Creator, msg.ProofType, msg.PublicParameters, msg.VerificationKey, msg.CircuitId)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgRegisterZKProofCircuitResponse{ProofId: proofID}, nil
}

func (ms msgServer) SubmitZKProof(goCtx context.Context, msg *cryptoproto.MsgSubmitZKProof) (*cryptoproto.MsgSubmitZKProofResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid submitter address")
	}

	if !signers[0].Equals(submitterAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match submitter")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.ProofId == "" || len(msg.ProofData) == 0 {
		return nil, types.ErrInvalidInput.Wrap("proof_id and proof_data are required")
	}

	// Verify that the proof circuit exists
	_, err = ms.Keeper.GetZKProofConfig(ctx, msg.ProofId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ZK proof circuit not found")
	}

	verified, verificationID, err := ms.Keeper.SubmitZKProof(ctx, msg.Submitter, msg.ProofId, msg.ProofData, msg.PublicInputs)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgSubmitZKProofResponse{
		Verified:       verified,
		VerificationId: verificationID,
	}, nil
}

func (ms msgServer) RegisterSecureEnclave(goCtx context.Context, msg *cryptoproto.MsgRegisterSecureEnclave) (*cryptoproto.MsgRegisterSecureEnclaveResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(msg.AttestationData) == 0 {
		return nil, types.ErrInvalidInput.Wrap("attestation_data is required")
	}

	enclaveID, err := ms.Keeper.RegisterSecureEnclave(ctx, msg.Creator, msg.EnclaveType, msg.AttestationData, msg.EnclaveMetadata)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgRegisterSecureEnclaveResponse{EnclaveId: enclaveID}, nil
}

func (ms msgServer) GenerateQuantumResistantKey(goCtx context.Context, msg *cryptoproto.MsgGenerateQuantumResistantKey) (*cryptoproto.MsgGenerateQuantumResistantKeyResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate that public key is provided (MUST be generated client-side)
	if len(msg.PublicKey) == 0 {
		return nil, status.Error(codes.InvalidArgument, "public_key is required (must be generated off-chain)")
	}

	// Register the quantum-resistant public key that was generated off-chain
	// msg.ExpiresAt is already *time.Time, no conversion needed
	keyID, err := ms.Keeper.RegisterQuantumResistantKey(ctx, msg.Creator, msg.Algorithm, msg.PublicKey, msg.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgGenerateQuantumResistantKeyResponse{
		KeyId: keyID,
	}, nil
}

func (ms msgServer) AddCertificatePin(goCtx context.Context, msg *cryptoproto.MsgAddCertificatePin) (*cryptoproto.MsgAddCertificatePinResponse, error) {
	// Verify signer
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match creator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Hostname == "" || len(msg.CertificateHashes) == 0 {
		return nil, types.ErrInvalidInput.Wrap("hostname and certificate_hashes are required")
	}

	// msg.ExpiresAt is already *time.Time, no conversion needed
	pinID, err := ms.Keeper.AddCertificatePin(ctx, msg.Creator, msg.Hostname, msg.CertificateHashes, msg.PinType, msg.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgAddCertificatePinResponse{PinId: pinID}, nil
}

func (ms msgServer) UpdateParams(goCtx context.Context, msg *cryptoproto.MsgUpdateParams) (*cryptoproto.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.UpdateParams(ctx, msg.Authority, &msg.Params); err != nil {
		return nil, err
	}

	return &cryptoproto.MsgUpdateParamsResponse{}, nil
}
