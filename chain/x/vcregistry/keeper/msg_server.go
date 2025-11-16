package keeper

import (
	"context"
	"fmt"
	"time"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MsgServer implements the Msg service
type MsgServer struct {
	vcregistrypb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer returns a new MsgServer
func NewMsgServer(keeper *Keeper) *MsgServer {
	return &MsgServer{keeper: keeper}
}

var _ vcregistrypb.MsgServer = &MsgServer{}

// ============================
// PRESENTATION MESSAGES
// ============================

// CreatePresentation handles MsgCreatePresentation
func (m *MsgServer) CreatePresentation(
	ctx context.Context,
	msg *vcregistrypb.MsgCreatePresentation,
) (*vcregistrypb.MsgCreatePresentationResponse, error) {
	// Validate inputs
	if msg.Creator == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if len(msg.VcIds) == 0 {
		return nil, types.ErrEmptyVCList
	}
	if msg.Context == nil {
		return nil, types.ErrInvalidInput
	}

	// Set default expiration if not provided
	expiresInSeconds := msg.ExpiresInSeconds
	if expiresInSeconds == 0 {
		expiresInSeconds = 300 // 5 minutes default
	}

	// Create presentation
	presentation, qrCodeData, err := m.keeper.CreatePresentation(
		msg.Creator,
		msg.VcIds,
		msg.Context,
		expiresInSeconds,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create presentation: %w", err)
	}

	// Emit event
	// TODO: Emit EventPresentationCreated event using SDK event manager

	return &vcregistrypb.MsgCreatePresentationResponse{
		PresentationId: presentation.PresentationId,
		QrCodeData:     qrCodeData,
		ExpiresAt:      presentation.CreatedAt, // Should be createdAt + expiresInSeconds
	}, nil
}

// Placeholder implementations for other message types
// These would be implemented based on existing vcregistry functionality

func (m *MsgServer) MintVC(
	ctx context.Context,
	msg *vcregistrypb.MsgMintVC,
) (*vcregistrypb.MsgMintVCResponse, error) {
	// Validate inputs
	if msg.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.HolderDid == "" {
		return nil, types.ErrInvalidDID
	}
	if msg.VcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidVCType
	}

	// Mint the VC
	vcID, err := m.keeper.MintVC(
		msg.HolderAddress,
		msg.HolderDid,
		msg.VcType,
		msg.VcTypeCustom,
		msg.Metadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to mint VC: %w", err)
	}

	// Get the created VC to return details
	vcRecord, ok := m.keeper.GetVCRecord(vcID)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// TODO: Emit EventVCMinted event using SDK event manager

	return &vcregistrypb.MsgMintVCResponse{
		VcId:           vcID,
		IssuedAt:       vcRecord.IssuedAt,
		ExpiresAt:      vcRecord.ExpiresAt,
		CredentialHash: vcRecord.CredentialHash,
	}, nil
}

func (m *MsgServer) RevokeVC(
	ctx context.Context,
	msg *vcregistrypb.MsgRevokeVC,
) (*vcregistrypb.MsgRevokeVCResponse, error) {
	// Validate inputs
	if msg.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify the signer owns the VC
	vcRecord, ok := m.keeper.GetVCRecord(msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}
	if vcRecord.HolderAddress != msg.HolderAddress {
		return nil, types.ErrNotVCHolder
	}

	// Revoke the VC (user-initiated revocations use USER_REQUEST reason)
	err := m.keeper.RevokeVC(
		msg.VcId,
		vcregistrypb.RevocationReason_REVOCATION_REASON_USER_REQUEST,
		msg.HolderAddress,
		msg.ReasonText,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke VC: %w", err)
	}

	// Get revocation record for response
	revRecord, _ := m.keeper.GetRevocationRecord(msg.VcId)

	// TODO: Emit EventVCRevoked event

	return &vcregistrypb.MsgRevokeVCResponse{
		RevokedAt:     revRecord.RevokedAt,
		MerkleUpdated: true,
	}, nil
}

func (m *MsgServer) AdminRevokeVC(
	ctx context.Context,
	msg *vcregistrypb.MsgAdminRevokeVC,
) (*vcregistrypb.MsgAdminRevokeVCResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Revoke the VC with admin reason
	err := m.keeper.RevokeVC(
		msg.VcId,
		msg.Reason,
		"governance",
		msg.Evidence,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to admin revoke VC: %w", err)
	}

	// Get revocation record
	revRecord, _ := m.keeper.GetRevocationRecord(msg.VcId)

	// TODO: Emit EventVCRevoked event

	return &vcregistrypb.MsgAdminRevokeVCResponse{
		RevokedAt:     revRecord.RevokedAt,
		MerkleUpdated: true,
	}, nil
}

func (m *MsgServer) SuspendVC(
	ctx context.Context,
	msg *vcregistrypb.MsgSuspendVC,
) (*vcregistrypb.MsgSuspendVCResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get VC record
	vcRecord, ok := m.keeper.GetVCRecord(msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// Check current status
	if vcRecord.Status == vcregistrypb.VCStatus_VC_STATUS_REVOKED {
		return nil, types.ErrVCAlreadyRevoked
	}
	if vcRecord.Status == vcregistrypb.VCStatus_VC_STATUS_SUSPENDED {
		return nil, types.ErrVCSuspended
	}

	// Update status to suspended
	vcRecord.Status = vcregistrypb.VCStatus_VC_STATUS_SUSPENDED
	if err := m.keeper.SetVCRecord(vcRecord); err != nil {
		return nil, fmt.Errorf("failed to suspend VC: %w", err)
	}

	// Calculate reactivation time if duration specified
	var reactivateAt *timestamppb.Timestamp
	if msg.SuspensionDurationDays > 0 {
		reactivateTime := m.keeper.currentTime + (int64(msg.SuspensionDurationDays) * 86400)
		reactivateAt = timestamppb.New(time.Unix(reactivateTime, 0))
	}

	// TODO: Emit EventVCSuspended event

	return &vcregistrypb.MsgSuspendVCResponse{
		SuspendedAt:  timestamppb.New(time.Unix(m.keeper.currentTime, 0)),
		ReactivateAt: reactivateAt,
	}, nil
}

func (m *MsgServer) ReactivateVC(
	ctx context.Context,
	msg *vcregistrypb.MsgReactivateVC,
) (*vcregistrypb.MsgReactivateVCResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get VC record
	vcRecord, ok := m.keeper.GetVCRecord(msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// Check current status
	if vcRecord.Status != vcregistrypb.VCStatus_VC_STATUS_SUSPENDED {
		return nil, fmt.Errorf("VC is not suspended (current status: %s)", vcRecord.Status.String())
	}

	// Update status to active
	vcRecord.Status = vcregistrypb.VCStatus_VC_STATUS_ACTIVE
	if err := m.keeper.SetVCRecord(vcRecord); err != nil {
		return nil, fmt.Errorf("failed to reactivate VC: %w", err)
	}

	// TODO: Emit EventVCReactivated event

	return &vcregistrypb.MsgReactivateVCResponse{
		ReactivatedAt: timestamppb.New(time.Unix(m.keeper.currentTime, 0)),
	}, nil
}

func (m *MsgServer) CreateVCPolicy(
	ctx context.Context,
	msg *vcregistrypb.MsgCreateVCPolicy,
) (*vcregistrypb.MsgCreateVCPolicyResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Check if policy already exists
	if _, ok := m.keeper.GetVCPolicy(msg.VcTypeName); ok {
		return nil, fmt.Errorf("policy already exists for type: %s", msg.VcTypeName)
	}

	// Create policy version
	version := "v1.0"

	// Create policy
	policy := vcregistrypb.VCPolicy{
		VcTypeName:            msg.VcTypeName,
		VcTypeEnum:            msg.VcTypeEnum,
		CsThreshold:           msg.CsThreshold,
		RequiredIrIds:         msg.RequiredIrIds,
		RequiredArena:         msg.RequiredArena,
		RequiredArenaScore:    msg.RequiredArenaScore,
		ExpiryDurationDays:    msg.ExpiryDurationDays,
		Singleton:             msg.Singleton,
		RequiresAnnualRenewal: msg.RequiresAnnualRenewal,
		MetadataUri:           msg.MetadataUri,
		Status:                vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:               version,
		CreatedAt:             timestamppb.New(time.Unix(m.keeper.currentTime, 0)),
		CreatedHeight:         m.keeper.currentHeight,
		Creator:               msg.Authority,
	}

	// Store policy
	if err := m.keeper.SetVCPolicy(policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	// TODO: Emit EventVCPolicyCreated event

	return &vcregistrypb.MsgCreateVCPolicyResponse{
		PolicyId: msg.VcTypeName,
		Version:  version,
	}, nil
}

func (m *MsgServer) UpdateVCPolicy(
	ctx context.Context,
	msg *vcregistrypb.MsgUpdateVCPolicy,
) (*vcregistrypb.MsgUpdateVCPolicyResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get existing policy
	existingPolicy, ok := m.keeper.GetVCPolicy(msg.VcTypeName)
	if !ok {
		return nil, types.ErrPolicyNotFound
	}

	// Check if policy is deprecated
	if existingPolicy.Status == vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED {
		return nil, types.ErrPolicyDeprecated
	}

	// Increment version (simple increment for now)
	oldVersion := existingPolicy.Version
	newVersion := fmt.Sprintf("v%d.0", len(oldVersion)+1) // Simple versioning

	// Update policy fields while preserving creation data
	existingPolicy.CsThreshold = msg.CsThreshold
	existingPolicy.RequiredIrIds = msg.RequiredIrIds
	existingPolicy.RequiredArena = msg.RequiredArena
	existingPolicy.RequiredArenaScore = msg.RequiredArenaScore
	existingPolicy.ExpiryDurationDays = msg.ExpiryDurationDays
	existingPolicy.Singleton = msg.Singleton
	existingPolicy.RequiresAnnualRenewal = msg.RequiresAnnualRenewal
	existingPolicy.MetadataUri = msg.MetadataUri
	existingPolicy.Version = newVersion

	// Store updated policy
	if err := m.keeper.SetVCPolicy(existingPolicy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	// TODO: Emit EventVCPolicyUpdated event

	return &vcregistrypb.MsgUpdateVCPolicyResponse{
		NewVersion: newVersion,
	}, nil
}

func (m *MsgServer) DeprecateVCPolicy(
	ctx context.Context,
	msg *vcregistrypb.MsgDeprecateVCPolicy,
) (*vcregistrypb.MsgDeprecateVCPolicyResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, types.ErrUnauthorized
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get existing policy
	existingPolicy, ok := m.keeper.GetVCPolicy(msg.VcTypeName)
	if !ok {
		return nil, types.ErrPolicyNotFound
	}

	// Check if already deprecated
	if existingPolicy.Status == vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED {
		return nil, types.ErrPolicyDeprecated
	}

	// Update status to deprecated
	existingPolicy.Status = vcregistrypb.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED

	// Store updated policy
	if err := m.keeper.SetVCPolicy(existingPolicy); err != nil {
		return nil, fmt.Errorf("failed to deprecate policy: %w", err)
	}

	deprecatedAt := timestamppb.New(time.Unix(m.keeper.currentTime, 0))

	// TODO: Emit EventVCPolicyDeprecated event

	return &vcregistrypb.MsgDeprecateVCPolicyResponse{
		DeprecatedAt: deprecatedAt,
	}, nil
}

func (m *MsgServer) RegisterDID(
	ctx context.Context,
	msg *vcregistrypb.MsgRegisterDID,
) (*vcregistrypb.MsgRegisterDIDResponse, error) {
	// Validate inputs
	if msg.Controller == "" {
		return nil, types.ErrInvalidController
	}
	if msg.Did == "" {
		return nil, types.ErrInvalidDID
	}

	// Register the DID
	err := m.keeper.RegisterDID(
		msg.Did,
		msg.Controller,
		msg.VerificationMethods,
		msg.MetadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register DID: %w", err)
	}

	createdAt := timestamppb.New(time.Unix(m.keeper.currentTime, 0))

	// TODO: Emit EventDIDRegistered event

	return &vcregistrypb.MsgRegisterDIDResponse{
		Did:       msg.Did,
		CreatedAt: createdAt,
	}, nil
}

func (m *MsgServer) UpdateDIDDocument(
	ctx context.Context,
	msg *vcregistrypb.MsgUpdateDIDDocument,
) (*vcregistrypb.MsgUpdateDIDDocumentResponse, error) {
	// Validate inputs
	if msg.Controller == "" {
		return nil, types.ErrInvalidController
	}
	if msg.Did == "" {
		return nil, types.ErrInvalidDID
	}

	// Get existing DID to verify controller
	existingDoc, ok := m.keeper.GetDIDDocument(msg.Did)
	if !ok {
		return nil, types.ErrDIDNotFound
	}

	// Verify the signer is the controller
	if existingDoc.Controller != msg.Controller {
		return nil, types.ErrInvalidController
	}

	// Update the DID document
	err := m.keeper.UpdateDIDDocument(
		msg.Did,
		msg.VerificationMethods,
		msg.MetadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update DID document: %w", err)
	}

	updatedAt := timestamppb.New(time.Unix(m.keeper.currentTime, 0))

	// TODO: Emit EventDIDUpdated event

	return &vcregistrypb.MsgUpdateDIDDocumentResponse{
		UpdatedAt: updatedAt,
	}, nil
}
