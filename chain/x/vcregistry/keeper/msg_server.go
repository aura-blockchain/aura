// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/vcregistry/types"
	vcregistrypb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
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

// timestampFromTime converts time.Time to gogotypes.Timestamp
func timestampFromTime(t time.Time) *gogotypes.Timestamp {
	return &gogotypes.Timestamp{Seconds: t.Unix(), Nanos: int32(t.Nanosecond())}
}

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

	// SECURITY: Verify transaction signer matches creator
	// This prevents attackers from creating presentations using other users' VCs
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("signer %s does not match creator %s", signers[0].String(), msg.Creator))
	}

	// Set default expiration if not provided
	expiresInSeconds := msg.ExpiresInSeconds
	if expiresInSeconds == 0 {
		expiresInSeconds = 300 // 5 minutes default
	}

	// Create presentation
	presentation, qrCodeData, err := m.keeper.CreatePresentation(
		ctx,
		msg.Creator,
		msg.VcIds,
		msg.Context,
		expiresInSeconds,
	)
	if err != nil {
		return nil, mapVCAuthorizationError(err)
	}

	// Emit event
	var presentationExpiresAt *gogotypes.Timestamp
	if presentation.CreatedAt != nil && presentation.ExpiresInSeconds > 0 {
		createdTime := time.Unix(presentation.CreatedAt.Seconds, int64(presentation.CreatedAt.Nanos))
		presentationExpiresAt = timestampFromTime(
			createdTime.Add(time.Duration(presentation.ExpiresInSeconds) * time.Second),
		)
	}

	sdkCtx := m.keeper.sdkContext(ctx)
	{
		attrs := map[string]string{
			"presentation_id": presentation.PresentationId,
			"holder_address":  msg.Creator,
			"vc_ids":          strings.Join(msg.VcIds, ","),
			"created_at":      formatTimestamp(presentation.CreatedAt),
			"expires_at":      formatTimestamp(presentationExpiresAt),
			"block_height":    fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		}
		m.emitEvent(ctx, types.EventTypePresentationCreated, attrs)
	}

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
	if msg == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if msg.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.HolderDid == "" {
		return nil, types.ErrInvalidDID
	}
	if msg.VcType == vcregistrypb.VCType_VC_TYPE_UNSPECIFIED {
		return nil, types.ErrInvalidVCType
	}

	// SECURITY: Verify transaction signer matches holder address
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	holderAddr, err := sdk.AccAddressFromBech32(msg.HolderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	if !signers[0].Equals(holderAddr) {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("signer %s does not match holder address %s", signers[0].String(), msg.HolderAddress))
	}

	// Mint the VC (cast vcregistrypb.VCType to types.VCType)
	vcID, err := m.keeper.MintVC(
		ctx,
		msg.HolderAddress,
		msg.HolderDid,
		types.VCType(msg.VcType),
		msg.VcTypeCustom,
		msg.Metadata,
	)
	if err != nil {
		return nil, mapVCAuthorizationError(err)
	}

	// Get the created VC to return details
	vcRecord, ok := m.keeper.GetVCRecord(ctx, vcID)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	sdkCtx := m.keeper.sdkContext(ctx)
	{
		attrs := types.NewEventVCMinted(
			vcID,
			vcRecord.VcType.String(),
			vcRecord.VcTypeCustom,
			vcRecord.HolderDid,
			vcRecord.HolderAddress,
			formatTimestamp(vcRecord.IssuedAt),
			formatTimestamp(vcRecord.ExpiresAt),
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
			vcRecord.PolicyVersion,
		)
		m.emitEvent(ctx, types.EventTypeVCMinted, attrs)
	}

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

	// SECURITY: Verify transaction signer matches holder address
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	holderAddr, err := sdk.AccAddressFromBech32(msg.HolderAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid holder address: %w", err)
	}

	if !signers[0].Equals(holderAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match holder address (%s)",
			signers[0].String(),
			msg.HolderAddress,
		))
	}

	// Verify the signer owns the VC
	vcRecord, ok := m.keeper.GetVCRecord(ctx, msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}
	if vcRecord.HolderAddress != msg.HolderAddress {
		return nil, types.ErrNotVCHolder
	}

	// Revoke the VC (user-initiated revocations use USER_REQUEST reason)
	err = m.keeper.RevokeVC(
		ctx,
		msg.VcId,
		types.RevocationReason(vcregistrypb.RevocationReason_REVOCATION_REASON_USER_REQUEST),
		msg.HolderAddress,
		msg.ReasonText,
	)
	if err != nil {
		return nil, mapVCAuthorizationError(err)
	}

	// Get revocation record for response
	revRecord, _ := m.keeper.GetRevocationRecord(ctx, msg.VcId)

	m.emitVCRevokedEvent(ctx, msg.VcId, vcRecord.VcType.String(), msg.ReasonText, msg.HolderAddress, revRecord.RevokedAt)

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
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
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
		ctx,
		msg.VcId,
		types.RevocationReason(msg.Reason),
		"governance",
		msg.Evidence,
	)
	if err != nil {
		return nil, mapVCAuthorizationError(err)
	}

	// Get revocation record
	revRecord, _ := m.keeper.GetRevocationRecord(ctx, msg.VcId)
	vcRecord, _ := m.keeper.GetVCRecord(ctx, msg.VcId)

	m.emitVCRevokedEvent(ctx, msg.VcId, vcRecord.VcType.String(), msg.Reason.String(), msg.Authority, revRecord.RevokedAt)

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
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority))
	}

	// Get VC record
	vcRecord, ok := m.keeper.GetVCRecord(ctx, msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// Check current status
	if vcRecord.Status == types.VCStatus_VC_STATUS_REVOKED {
		return nil, types.ErrVCAlreadyRevoked
	}
	if vcRecord.Status == types.VCStatus_VC_STATUS_SUSPENDED {
		return nil, types.ErrVCSuspended
	}

	// Update status to suspended
	vcRecord.Status = types.VCStatus_VC_STATUS_SUSPENDED
	if err := m.keeper.SetVCRecord(ctx, vcRecord); err != nil {
		return nil, fmt.Errorf("failed to suspend VC: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	suspendedAt := timestampFromTime(time.Unix(currentTime, 0))

	// Calculate reactivation time if duration specified
	var reactivateAt *gogotypes.Timestamp
	if msg.SuspensionDurationDays > 0 {
		reactivateTime := currentTime + (int64(msg.SuspensionDurationDays) * 86400)
		reactivateAt = timestampFromTime(time.Unix(reactivateTime, 0))
	}

	m.emitEvent(ctx, types.EventTypeVCSuspended, types.NewEventVCSuspended(
		msg.VcId,
		msg.Reason,
		formatTimestamp(suspendedAt),
		formatTimestamp(reactivateAt),
	))

	return &vcregistrypb.MsgSuspendVCResponse{
		SuspendedAt:  suspendedAt,
		ReactivateAt: reactivateAt,
	}, nil
}

func (m *MsgServer) ReactivateVC(
	ctx context.Context,
	msg *vcregistrypb.MsgReactivateVC,
) (*vcregistrypb.MsgReactivateVCResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
	}
	if msg.VcId == "" {
		return nil, types.ErrInvalidVCID
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get VC record
	vcRecord, ok := m.keeper.GetVCRecord(ctx, msg.VcId)
	if !ok {
		return nil, types.ErrVCNotFound
	}

	// Check current status
	if vcRecord.Status != types.VCStatus_VC_STATUS_SUSPENDED {
		return nil, fmt.Errorf("VC is not suspended (current status: %s)", vcRecord.Status.String())
	}

	// Update status to active
	vcRecord.Status = types.VCStatus_VC_STATUS_ACTIVE
	if err := m.keeper.SetVCRecord(ctx, vcRecord); err != nil {
		return nil, fmt.Errorf("failed to reactivate VC: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	reactivatedAt := timestampFromTime(time.Unix(currentTime, 0))

	m.emitEvent(ctx, types.EventTypeVCReactivated, types.NewEventVCReactivated(
		msg.VcId,
		formatTimestamp(reactivatedAt),
	))

	return &vcregistrypb.MsgReactivateVCResponse{
		ReactivatedAt: reactivatedAt,
	}, nil
}

func (m *MsgServer) CreateVCPolicy(
	ctx context.Context,
	msg *vcregistrypb.MsgCreateVCPolicy,
) (*vcregistrypb.MsgCreateVCPolicyResponse, error) {
	// Validate inputs
	if msg.Authority == "" {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Check if policy already exists
	if _, ok := m.keeper.GetVCPolicy(ctx, msg.VcTypeName); ok {
		return nil, fmt.Errorf("policy already exists for type: %s", msg.VcTypeName)
	}

	// Create policy version
	version := "v1.0"

	// Create policy (convert protobuf types to keeper types)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	currentHeight := uint64(sdkCtx.BlockHeight())

	policy := types.VCPolicy{
		VcTypeName:            msg.VcTypeName,
		VcTypeEnum:            types.VCType(msg.VcTypeEnum),
		CsThreshold:           msg.CsThreshold,
		RequiredIrIds:         msg.RequiredIrIds,
		RequiredArena:         msg.RequiredArena,
		RequiredArenaScore:    msg.RequiredArenaScore,
		ExpiryDurationDays:    msg.ExpiryDurationDays,
		Singleton:             msg.Singleton,
		RequiresAnnualRenewal: msg.RequiresAnnualRenewal,
		MetadataUri:           msg.MetadataUri,
		Status:                types.VCPolicyStatus_VC_POLICY_STATUS_ACTIVE,
		Version:               version,
		CreatedAt:             timestampFromTime(time.Unix(currentTime, 0)),
		CreatedHeight:         currentHeight,
		Creator:               msg.Authority,
	}

	// Store policy
	if err := m.keeper.SetVCPolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	{
		attrs := types.NewEventVCPolicyCreated(
			msg.VcTypeName,
			fmt.Sprintf("%d", msg.VcTypeEnum),
			fmt.Sprintf("%d", msg.CsThreshold),
			version,
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		)
		m.emitEvent(ctx, types.EventTypeVCPolicyCreated, attrs)
	}

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
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get existing policy
	existingPolicy, ok := m.keeper.GetVCPolicy(ctx, msg.VcTypeName)
	if !ok {
		return nil, types.ErrPolicyNotFound
	}

	// Check if policy is deprecated
	if existingPolicy.Status == types.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED {
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
	if err := m.keeper.SetVCPolicy(ctx, existingPolicy); err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	sdkCtx := m.keeper.sdkContext(ctx)
	{
		attrs := types.NewEventVCPolicyUpdated(
			msg.VcTypeName,
			oldVersion,
			newVersion,
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		)
		m.emitEvent(ctx, types.EventTypeVCPolicyUpdated, attrs)
	}

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
		return nil, mapVCAuthorizationError(types.ErrUnauthorized)
	}
	if msg.VcTypeName == "" {
		return nil, types.ErrInvalidVCType
	}

	// Verify authority is governance
	if msg.Authority != m.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", m.keeper.GetAuthority(), msg.Authority)
	}

	// Get existing policy
	existingPolicy, ok := m.keeper.GetVCPolicy(ctx, msg.VcTypeName)
	if !ok {
		return nil, types.ErrPolicyNotFound
	}

	// Check if already deprecated
	if existingPolicy.Status == types.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED {
		return nil, types.ErrPolicyDeprecated
	}

	// Update status to deprecated
	existingPolicy.Status = types.VCPolicyStatus_VC_POLICY_STATUS_DEPRECATED

	// Store updated policy
	if err := m.keeper.SetVCPolicy(ctx, existingPolicy); err != nil {
		return nil, fmt.Errorf("failed to deprecate policy: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	deprecatedAt := timestampFromTime(time.Unix(currentTime, 0))

	m.emitEvent(ctx, types.EventTypeVCPolicyDeprecated, types.NewEventVCPolicyDeprecated(
		msg.VcTypeName,
		msg.Reason,
		formatTimestamp(deprecatedAt),
	))

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

	// SECURITY: Verify transaction signer matches controller
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	controllerAddr, err := sdk.AccAddressFromBech32(msg.Controller)
	if err != nil {
		return nil, fmt.Errorf("invalid controller address: %w", err)
	}

	if !signers[0].Equals(controllerAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match controller (%s)",
			signers[0].String(),
			msg.Controller,
		))
	}

	// Register the DID (convert verification methods from protobuf to types)
	verificationMethods := types.VerificationMethodsFromProto(msg.VerificationMethods)
	err = m.keeper.RegisterDID(
		ctx,
		msg.Did,
		msg.Controller,
		verificationMethods,
		msg.MetadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register DID: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	createdAt := timestampFromTime(time.Unix(currentTime, 0))
	{
		attrs := types.NewEventDIDRegistered(
			msg.Did,
			msg.Controller,
			formatTimestamp(createdAt),
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		)
		m.emitEvent(ctx, types.EventTypeDIDRegistered, attrs)
	}

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

	// SECURITY: Verify transaction signer matches controller
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	controllerAddr, err := sdk.AccAddressFromBech32(msg.Controller)
	if err != nil {
		return nil, fmt.Errorf("invalid controller address: %w", err)
	}

	if !signers[0].Equals(controllerAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match controller (%s)",
			signers[0].String(),
			msg.Controller,
		))
	}

	// Get existing DID to verify controller
	existingDoc, ok := m.keeper.GetDIDDocument(ctx, msg.Did)
	if !ok {
		return nil, types.ErrDIDNotFound
	}

	// Verify the signer is the controller
	if existingDoc.Controller != msg.Controller {
		return nil, types.ErrInvalidController
	}

	// Update the DID document (convert verification methods from protobuf to types)
	updateVerificationMethods := types.VerificationMethodsFromProto(msg.VerificationMethods)
	err = m.keeper.UpdateDIDDocument(
		ctx,
		msg.Did,
		updateVerificationMethods,
		msg.MetadataUri,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update DID document: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	updatedAt := timestampFromTime(time.Unix(currentTime, 0))
	{
		attrs := types.NewEventDIDUpdated(
			msg.Did,
			formatTimestamp(updatedAt),
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		)
		m.emitEvent(ctx, types.EventTypeDIDUpdated, attrs)
	}

	return &vcregistrypb.MsgUpdateDIDDocumentResponse{
		UpdatedAt: updatedAt,
	}, nil
}

func (m *MsgServer) emitEvent(ctx context.Context, eventType string, attrs map[string]string) {
	sdkCtx := m.keeper.sdkContext(ctx)

	// Sort keys for deterministic iteration order (consensus-critical)
	keys := make([]string, 0, len(attrs))
	for key := range attrs {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	eventAttrs := make([]sdk.Attribute, 0, len(attrs))
	for _, key := range keys {
		value := attrs[key]
		if value == "" {
			continue
		}
		eventAttrs = append(eventAttrs, sdk.NewAttribute(key, value))
	}

	if len(eventAttrs) == 0 {
		return
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(eventType, eventAttrs...))
}

func formatTimestamp(ts *gogotypes.Timestamp) string {
	if ts == nil {
		return ""
	}
	return time.Unix(ts.Seconds, int64(ts.Nanos)).UTC().Format(time.RFC3339Nano)
}

func (m *MsgServer) emitVCRevokedEvent(ctx context.Context, vcID, vcType, reason, revoker string, revokedAt *gogotypes.Timestamp) {
	sdkCtx := m.keeper.sdkContext(ctx)
	{
		attrs := types.NewEventVCRevoked(
			vcID,
			vcType,
			reason,
			revoker,
			formatTimestamp(revokedAt),
			fmt.Sprintf("%d", sdkCtx.BlockHeight()),
		)
		m.emitEvent(ctx, types.EventTypeVCRevoked, attrs)
	}
}

func mapVCAuthorizationError(err error) error {
	if err == nil {
		return nil
	}

	if errorsmod.IsOf(err, types.ErrUnauthorized) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	return fmt.Errorf("error in mapVCAuthorizationError: %w", err)
}

// ============================
// ATTRIBUTE VC MESSAGES
// ============================

// CreateAttributeVC handles MsgCreateAttributeVC
func (m *MsgServer) CreateAttributeVC(
	ctx context.Context,
	msg *vcregistrypb.MsgCreateAttributeVC,
) (*vcregistrypb.MsgCreateAttributeVCResponse, error) {
	// Validate inputs
	if msg.Creator == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.AttributeType == vcregistrypb.AttributeType_ATTRIBUTE_TYPE_UNSPECIFIED {
		return nil, fmt.Errorf("attribute_type required")
	}
	if len(msg.EncryptedValue) == 0 {
		return nil, fmt.Errorf("encrypted_value required")
	}

	// SECURITY: Verify transaction signer matches creator
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match creator (%s)",
			signers[0].String(),
			msg.Creator,
		))
	}

	// Generate attribute VC ID
	avcID := m.keeper.GenerateAttributeVCID(ctx, msg.Creator, msg.AttributeType)

	// Calculate expiration
	var expiresAt *gogotypes.Timestamp
	if msg.ExpiresInSeconds > 0 {
		expirationTime := time.Unix(m.keeper.getCurrentTime(ctx), 0).Add(time.Duration(msg.ExpiresInSeconds) * time.Second)
		expiresAt = timestampFromTime(expirationTime)
	}

	// Create attribute VC
	avc := types.AttributeVC{
		AttributeVcId:  avcID,
		AttributeType:  msg.AttributeType,
		HolderAddress:  msg.Creator,
		EncryptedValue: msg.EncryptedValue,
		Issuer:         msg.Issuer,
		Status:         types.VCStatus_VC_STATUS_ACTIVE,
		IssuedAt:       timestampFromTime(time.Unix(m.keeper.getCurrentTime(ctx), 0)),
		ExpiresAt:      expiresAt,
	}

	if err := m.keeper.CreateAttributeVC(ctx, avc); err != nil {
		return nil, fmt.Errorf("failed to create attribute VC: %w", err)
	}

	// Emit event using standard pattern
	sdkCtx := m.keeper.sdkContext(ctx)
	m.emitEvent(ctx, "attribute_vc_created", map[string]string{
		"attribute_vc_id": avcID,
		"attribute_type":  msg.AttributeType.String(),
		"holder_address":  msg.Creator,
		"issuer":          msg.Issuer,
		"issued_at":       formatTimestamp(avc.IssuedAt),
		"block_height":    fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	})

	return &vcregistrypb.MsgCreateAttributeVCResponse{
		AttributeVcId: avcID,
		IssuedAt:      avc.IssuedAt,
		ExpiresAt:     expiresAt,
	}, nil
}

// RevokeAttributeVC handles MsgRevokeAttributeVC
func (m *MsgServer) RevokeAttributeVC(
	ctx context.Context,
	msg *vcregistrypb.MsgRevokeAttributeVC,
) (*vcregistrypb.MsgRevokeAttributeVCResponse, error) {
	// Validate inputs
	if msg.Creator == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.AttributeVcId == "" {
		return nil, fmt.Errorf("attribute_vc_id required")
	}

	// SECURITY: Verify transaction signer matches creator
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match creator (%s)",
			signers[0].String(),
			msg.Creator,
		))
	}

	// Verify the signer owns the attribute VC
	avc, ok := m.keeper.GetAttributeVC(ctx, msg.AttributeVcId)
	if !ok {
		return nil, fmt.Errorf("attribute VC not found")
	}
	if avc.HolderAddress != msg.Creator {
		return nil, types.ErrNotVCHolder
	}

	// Revoke the attribute VC
	if err := m.keeper.RevokeAttributeVC(ctx, msg.AttributeVcId, msg.Reason); err != nil {
		return nil, fmt.Errorf("failed to revoke attribute VC: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime().Unix()
	revokedAt := timestampFromTime(time.Unix(currentTime, 0))

	m.emitEvent(ctx, "attribute_vc_revoked", map[string]string{
		"attribute_vc_id": msg.AttributeVcId,
		"attribute_type":  avc.AttributeType.String(),
		"reason":          msg.Reason,
		"revoked_at":      formatTimestamp(revokedAt),
		"block_height":    fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	})

	return &vcregistrypb.MsgRevokeAttributeVCResponse{
		RevokedAt: revokedAt,
	}, nil
}

// UpdateDisclosurePolicy handles MsgUpdateDisclosurePolicy
func (m *MsgServer) UpdateDisclosurePolicy(
	ctx context.Context,
	msg *vcregistrypb.MsgUpdateDisclosurePolicy,
) (*vcregistrypb.MsgUpdateDisclosurePolicyResponse, error) {
	// Validate inputs
	if msg.Creator == "" {
		return nil, types.ErrInvalidHolderAddress
	}

	// SECURITY: Verify transaction signer matches creator
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match creator (%s)",
			signers[0].String(),
			msg.Creator,
		))
	}

	// Create policy
	policy := types.DisclosurePolicy{
		HolderAddress: msg.Creator,
		DefaultMode:   msg.DefaultMode,
		Rules:         msg.Rules,
		UpdatedAt:     timestampFromTime(time.Unix(m.keeper.getCurrentTime(ctx), 0)),
	}

	if err := m.keeper.SetDisclosurePolicy(ctx, policy); err != nil {
		return nil, fmt.Errorf("failed to set disclosure policy: %w", err)
	}

	sdkCtx := m.keeper.sdkContext(ctx)
	m.emitEvent(ctx, "disclosure_policy_updated", map[string]string{
		"holder_address": msg.Creator,
		"default_mode":   msg.DefaultMode.String(),
		"rules_count":    fmt.Sprintf("%d", len(msg.Rules)),
		"updated_at":     formatTimestamp(policy.UpdatedAt),
		"block_height":   fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	})

	return &vcregistrypb.MsgUpdateDisclosurePolicyResponse{
		UpdatedAt: policy.UpdatedAt,
	}, nil
}

// ============================
// DISCLOSURE REQUEST/RESPONSE MESSAGES
// ============================

// CreateDisclosureRequest handles MsgCreateDisclosureRequest
func (m *MsgServer) CreateDisclosureRequest(
	ctx context.Context,
	msg *vcregistrypb.MsgCreateDisclosureRequest,
) (*vcregistrypb.MsgCreateDisclosureRequestResponse, error) {
	// Validate inputs
	if msg.HolderAddress == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.Verifier == "" {
		return nil, fmt.Errorf("verifier required")
	}
	if len(msg.RequestedAttributes) == 0 {
		return nil, fmt.Errorf("requested_attributes required")
	}

	// Generate request ID
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	requestID := fmt.Sprintf("disc-req-%s-%d-%d",
		msg.HolderAddress,
		sdkCtx.BlockHeight(),
		sdkCtx.BlockTime().Unix(),
	)

	// Create disclosure request
	req := types.DisclosureRequest{
		RequestId:           requestID,
		VerifierAddress:     msg.Verifier,
		VerifierName:        msg.VerifierName,
		RequestedAttributes: msg.RequestedAttributes,
		Purpose:             msg.Purpose,
		ExpiresInSeconds:    msg.ExpiresInSeconds,
		RequestedAt:         timestampFromTime(sdkCtx.BlockTime()),
	}

	if err := m.keeper.CreateDisclosureRequest(ctx, msg.HolderAddress, req); err != nil {
		return nil, fmt.Errorf("failed to create disclosure request: %w", err)
	}

	// Calculate expiration time
	expiresAt := timestampFromTime(sdkCtx.BlockTime().Add(time.Duration(msg.ExpiresInSeconds) * time.Second))

	m.emitEvent(ctx, "disclosure_request_created", map[string]string{
		"request_id":       requestID,
		"holder_address":   msg.HolderAddress,
		"verifier":         msg.Verifier,
		"attributes_count": fmt.Sprintf("%d", len(msg.RequestedAttributes)),
		"requested_at":     formatTimestamp(req.RequestedAt),
		"block_height":     fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	})

	return &vcregistrypb.MsgCreateDisclosureRequestResponse{
		RequestId: requestID,
		ExpiresAt: expiresAt,
	}, nil
}

// RespondToDisclosureRequest handles MsgRespondToDisclosureRequest
func (m *MsgServer) RespondToDisclosureRequest(
	ctx context.Context,
	msg *vcregistrypb.MsgRespondToDisclosureRequest,
) (*vcregistrypb.MsgRespondToDisclosureRequestResponse, error) {
	// Validate inputs
	if msg.Creator == "" {
		return nil, types.ErrInvalidHolderAddress
	}
	if msg.RequestId == "" {
		return nil, fmt.Errorf("request_id required")
	}

	// SECURITY: Verify transaction signer matches creator (holder responding)
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, fmt.Errorf("no signers")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, fmt.Errorf("invalid creator address: %w", err)
	}

	if !signers[0].Equals(creatorAddr) {
		return nil, mapVCAuthorizationError(types.ErrUnauthorized.Wrapf(
			"signer (%s) does not match creator (%s)",
			signers[0].String(),
			msg.Creator,
		))
	}

	// Convert AttributeType list to AttributeDisclosure list
	// If approved, we need to fetch the actual attribute VCs and create disclosures
	var disclosedAttrs []*vcregistrypb.AttributeDisclosure
	if msg.Approved && len(msg.DisclosedAttributes) > 0 {
		// For each disclosed attribute type, create a disclosure entry
		// Note: In production, this would decrypt the value or generate a ZK proof
		for _, attrType := range msg.DisclosedAttributes {
			// Get attribute VCs for this holder
			avcs := m.keeper.ListAttributeVCs(ctx, msg.Creator, []types.AttributeType{attrType})
			if len(avcs) > 0 {
				// Use the first active VC of this type
				// In production, would decrypt EncryptedValue to RevealedValue
				disclosedAttrs = append(disclosedAttrs, &vcregistrypb.AttributeDisclosure{
					AttributeType: attrType,
					RevealedValue: "<encrypted>", // Placeholder - would decrypt in production
					IsZkProof:     false,
				})
			}
		}
	}

	// Create response
	resp := types.DisclosureResponse{
		RequestId:           msg.RequestId,
		HolderAddress:       msg.Creator,
		Approved:            msg.Approved,
		DisclosedAttributes: disclosedAttrs,
		RespondedAt:         timestampFromTime(time.Unix(m.keeper.getCurrentTime(ctx), 0)),
	}

	if err := m.keeper.RespondToDisclosureRequest(ctx, resp); err != nil {
		return nil, fmt.Errorf("failed to respond to disclosure request: %w", err)
	}

	sdkCtx := m.keeper.sdkContext(ctx)
	m.emitEvent(ctx, "disclosure_response_created", map[string]string{
		"request_id":       msg.RequestId,
		"holder_address":   msg.Creator,
		"approved":         fmt.Sprintf("%t", msg.Approved),
		"attributes_count": fmt.Sprintf("%d", len(disclosedAttrs)),
		"responded_at":     formatTimestamp(resp.RespondedAt),
		"block_height":     fmt.Sprintf("%d", sdkCtx.BlockHeight()),
	})

	return &vcregistrypb.MsgRespondToDisclosureRequestResponse{
		Response: &resp,
	}, nil
}
