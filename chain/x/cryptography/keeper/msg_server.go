package keeper

import (
	"context"

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.KeyId == "" || len(msg.NewPublicKey) == 0 {
		return nil, types.ErrInvalidInput.Wrap("key_id and new_public_key are required")
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Threshold <= 0 || msg.TotalParticipants <= 0 || msg.Threshold > msg.TotalParticipants {
		return nil, types.ErrInvalidInput.Wrap("invalid threshold or participant count")
	}

	if len(msg.ParticipantIds) != int(msg.TotalParticipants) {
		return nil, types.ErrInvalidInput.Wrap("participant_ids count must match total_participants")
	}

	schemeID, publicKey, err := ms.Keeper.CreateThresholdScheme(ctx, msg.Creator, msg.Threshold, msg.TotalParticipants, msg.ParticipantIds, msg.SchemeType)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgCreateThresholdSchemeResponse{
		SchemeId:  schemeID,
		PublicKey: publicKey,
	}, nil
}

func (ms msgServer) SubmitThresholdSignatureShare(goCtx context.Context, msg *cryptoproto.MsgSubmitThresholdSignatureShare) (*cryptoproto.MsgSubmitThresholdSignatureShareResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.SchemeId == "" || len(msg.SignatureShare) == 0 || len(msg.MessageHash) == 0 {
		return nil, types.ErrInvalidInput.Wrap("scheme_id, signature_share, and message_hash are required")
	}

	sharesCollected, thresholdReached, combinedSignature, err := ms.Keeper.SubmitThresholdSignatureShare(ctx, msg.Submitter, msg.SchemeId, msg.SignatureShare, msg.MessageHash)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgSubmitThresholdSignatureShareResponse{
		SharesCollected:   sharesCollected,
		ThresholdReached:  thresholdReached,
		CombinedSignature: combinedSignature,
	}, nil
}

func (ms msgServer) RegisterZKProofCircuit(goCtx context.Context, msg *cryptoproto.MsgRegisterZKProofCircuit) (*cryptoproto.MsgRegisterZKProofCircuitResponse, error) {
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.ProofId == "" || len(msg.ProofData) == 0 {
		return nil, types.ErrInvalidInput.Wrap("proof_id and proof_data are required")
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	keyID, publicKey, err := ms.Keeper.GenerateQuantumResistantKey(ctx, msg.Creator, msg.Algorithm, msg.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &cryptoproto.MsgGenerateQuantumResistantKeyResponse{
		KeyId:     keyID,
		PublicKey: publicKey,
	}, nil
}

func (ms msgServer) AddCertificatePin(goCtx context.Context, msg *cryptoproto.MsgAddCertificatePin) (*cryptoproto.MsgAddCertificatePinResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Hostname == "" || len(msg.CertificateHashes) == 0 {
		return nil, types.ErrInvalidInput.Wrap("hostname and certificate_hashes are required")
	}

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
