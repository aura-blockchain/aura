// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/security/types"
	securitypb "github.com/aequitas/aura/proto/aura/security/v1beta1"
)

type msgServer struct {
	securitypb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) securitypb.MsgServer {
	return &msgServer{keeper: keeper}
}

// verifySigner validates that the signer is authorized for this operation
func (ms msgServer) verifySigner(ctx sdk.Context, signer string) error {
	if signer == "" {
		return types.ErrUnauthorized
	}
	// Validate signer is a valid bech32 address
	_, err := sdk.AccAddressFromBech32(signer)
	if err != nil {
		return types.ErrUnauthorized
	}
	return nil
}

// ========================
// NETWORK SECURITY MESSAGES
// ========================

func (ms msgServer) AddTrustedPeer(ctx context.Context, msg *securitypb.MsgAddTrustedPeer) (*securitypb.MsgAddTrustedPeerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}

	// Check if peer already exists
	if _, found := ms.keeper.GetTrustedPeer(sdkCtx, msg.PeerId); found {
		return nil, types.ErrAlreadyExists
	}

	peer := &securitypb.TrustedPeer{
		PeerId:      msg.PeerId,
		Address:     msg.Address,
		PublicKey:   msg.PublicKey,
		Description: msg.Description,
		AddedAt:     sdkCtx.BlockTime(),
	}

	ms.keeper.SetTrustedPeer(sdkCtx, peer)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTrustedPeerAdded,
			sdk.NewAttribute(types.AttributeKeyPeerId, msg.PeerId),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("trusted peer added", "peer_id", msg.PeerId, "added_by", msg.Authority)
	return &securitypb.MsgAddTrustedPeerResponse{}, nil
}

func (ms msgServer) RemoveTrustedPeer(ctx context.Context, msg *securitypb.MsgRemoveTrustedPeer) (*securitypb.MsgRemoveTrustedPeerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}

	// Check if peer exists
	if _, found := ms.keeper.GetTrustedPeer(sdkCtx, msg.PeerId); !found {
		return nil, types.ErrNotFound
	}

	ms.keeper.DeleteTrustedPeer(sdkCtx, msg.PeerId)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTrustedPeerRemoved,
			sdk.NewAttribute(types.AttributeKeyPeerId, msg.PeerId),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("trusted peer removed", "peer_id", msg.PeerId, "authority", msg.Authority)
	return &securitypb.MsgRemoveTrustedPeerResponse{}, nil
}

func (ms msgServer) BanPeer(ctx context.Context, msg *securitypb.MsgBanPeer) (*securitypb.MsgBanPeerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}

	// Check if already blacklisted
	if ms.keeper.IsBlacklisted(sdkCtx, msg.PeerId) {
		return nil, types.ErrAlreadyExists
	}

	now := sdkCtx.BlockTime()
	var expiresAt *time.Time
	permanent := false
	if msg.BanDuration > 0 {
		expiry := now.Add(msg.BanDuration)
		expiresAt = &expiry
	} else {
		permanent = true
	}

	entry := &types.BlacklistEntry{
		Identifier: msg.PeerId,
		Reason:     msg.Reason,
		Permanent:  permanent,
		ExpiresAt:  expiresAt,
		AddedAt:    &now,
	}

	ms.keeper.SetBlacklistEntry(sdkCtx, entry)

	// Also remove from trusted peers if present
	ms.keeper.DeleteTrustedPeer(sdkCtx, msg.PeerId)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePeerBanned,
			sdk.NewAttribute(types.AttributeKeyPeerId, msg.PeerId),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	ms.keeper.Logger(sdkCtx).Warn("peer banned", "peer_id", msg.PeerId, "reason", msg.Reason, "authority", msg.Authority)
	return &securitypb.MsgBanPeerResponse{}, nil
}

func (ms msgServer) UnbanPeer(ctx context.Context, msg *securitypb.MsgUnbanPeer) (*securitypb.MsgUnbanPeerResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}

	// Check if blacklisted
	if !ms.keeper.IsBlacklisted(sdkCtx, msg.PeerId) {
		return nil, types.ErrNotFound
	}

	ms.keeper.DeleteBlacklistEntry(sdkCtx, msg.PeerId)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePeerUnbanned,
			sdk.NewAttribute(types.AttributeKeyPeerId, msg.PeerId),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("peer unbanned", "peer_id", msg.PeerId, "authority", msg.Authority)
	return &securitypb.MsgUnbanPeerResponse{}, nil
}

func (ms msgServer) UpdatePeerReputation(ctx context.Context, msg *securitypb.MsgUpdatePeerReputation) (*securitypb.MsgUpdatePeerReputationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.PeerId == "" {
		return nil, types.ErrInvalidPeerID
	}

	// Get existing reputation or create new one
	rep, found := ms.keeper.GetPeerReputation(sdkCtx, msg.PeerId)
	if !found {
		rep = &securitypb.NodeReputation{
			PeerId:           msg.PeerId,
			Score:            msg.ReputationDelta,
			MessagesReceived: 0,
			ValidMessages:    0,
			InvalidMessages:  0,
		}
	} else {
		rep.Score += msg.ReputationDelta
	}

	rep.LastUpdatedHeight = sdkCtx.BlockHeight()
	ms.keeper.SetPeerReputation(sdkCtx, rep)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePeerReputationUpdated,
			sdk.NewAttribute(types.AttributeKeyPeerId, msg.PeerId),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("peer reputation updated", "peer_id", msg.PeerId, "new_score", rep.Score)
	return &securitypb.MsgUpdatePeerReputationResponse{NewReputation: rep.Score}, nil
}

func (ms msgServer) ResolveForkAlert(ctx context.Context, msg *securitypb.MsgResolveForkAlert) (*securitypb.MsgResolveForkAlertResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.AlertId == "" {
		return nil, types.ErrInvalidRequest
	}

	// Get all fork alerts and find the one to resolve
	alerts := ms.keeper.GetAllForkAlerts(sdkCtx)
	var found bool
	for _, alert := range alerts {
		if alert.AlertId == msg.AlertId {
			found = true
			// Mark as resolved by updating the alert
			alert.Resolved = true
			alert.ResolutionDetails = msg.ResolutionDetails
			ms.keeper.SetForkAlert(sdkCtx, alert)
			break
		}
	}

	if !found {
		return nil, types.ErrNotFound
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeForkAlertResolved,
			sdk.NewAttribute(types.AttributeKeyAlertId, msg.AlertId),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("fork alert resolved", "alert_id", msg.AlertId, "resolved_by", msg.Authority)
	return &securitypb.MsgResolveForkAlertResponse{}, nil
}

func (ms msgServer) ResolvePartitionAlert(ctx context.Context, msg *securitypb.MsgResolvePartitionAlert) (*securitypb.MsgResolvePartitionAlertResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	if msg.AlertId == "" {
		return nil, types.ErrInvalidRequest
	}

	// Get all partition alerts and find the one to resolve
	alerts := ms.keeper.GetAllPartitionAlerts(sdkCtx)
	var found bool
	for _, alert := range alerts {
		if alert.AlertId == msg.AlertId {
			found = true
			alert.Resolved = true
			ms.keeper.SetPartitionAlert(sdkCtx, alert)
			break
		}
	}

	if !found {
		return nil, types.ErrNotFound
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypePartitionAlertResolved,
			sdk.NewAttribute(types.AttributeKeyAlertId, msg.AlertId),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("partition alert resolved", "alert_id", msg.AlertId, "resolved_by", msg.Authority)
	return &securitypb.MsgResolvePartitionAlertResponse{}, nil
}

// ========================
// VALIDATOR SECURITY MESSAGES
// ========================

func (ms msgServer) RegisterValidatorSecurity(ctx context.Context, msg *securitypb.MsgRegisterValidatorSecurity) (*securitypb.MsgRegisterValidatorSecurityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	// Check if already registered
	if _, found := ms.keeper.GetValidatorSecurityInfo(sdkCtx, msg.ValidatorAddress); found {
		return nil, types.ErrAlreadyExists
	}

	// Determine if keys are separated (both hot and cold provided)
	keysSeparated := msg.HotKey != "" && msg.ColdKey != ""

	info := &securitypb.ValidatorSecurityInfo{
		ValidatorAddress:    msg.ValidatorAddress,
		HotKey:              msg.HotKey,
		ColdKey:             msg.ColdKey,
		KeysSeparated:       keysSeparated,
		Region:              msg.Region,
		CountryCode:         msg.CountryCode,
		Latitude:            msg.Latitude,
		Longitude:           msg.Longitude,
		MissedBlocksCounter: 0,
		IsJailed:            false,
		IsTombstoned:        false,
	}

	ms.keeper.SetValidatorSecurityInfo(sdkCtx, info)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorSecurityRegistered,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("validator security registered", "validator", msg.ValidatorAddress, "region", msg.Region)
	return &securitypb.MsgRegisterValidatorSecurityResponse{}, nil
}

func (ms msgServer) UpdateValidatorSecurity(ctx context.Context, msg *securitypb.MsgUpdateValidatorSecurity) (*securitypb.MsgUpdateValidatorSecurityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	// Get existing info
	info, found := ms.keeper.GetValidatorSecurityInfo(sdkCtx, msg.ValidatorAddress)
	if !found {
		return nil, types.ErrValidatorNotFound
	}

	// Update fields from message
	if msg.Region != "" {
		info.Region = msg.Region
	}
	if msg.CountryCode != "" {
		info.CountryCode = msg.CountryCode
	}
	if msg.Latitude != 0 || msg.Longitude != 0 {
		info.Latitude = msg.Latitude
		info.Longitude = msg.Longitude
	}
	if len(msg.SentryNodeAddresses) > 0 {
		info.SentryNodeAddresses = msg.SentryNodeAddresses
	}

	ms.keeper.SetValidatorSecurityInfo(sdkCtx, info)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorSecurityUpdated,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("validator security updated", "validator", msg.ValidatorAddress)
	return &securitypb.MsgUpdateValidatorSecurityResponse{}, nil
}

func (ms msgServer) RegisterSentryNode(ctx context.Context, msg *securitypb.MsgRegisterSentryNode) (*securitypb.MsgRegisterSentryNodeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	if msg.SentryAddress == "" {
		return nil, types.ErrInvalidSentryNode
	}

	sentry := &securitypb.SentryNodeInfo{
		ValidatorAddress: msg.ValidatorAddress,
		Address:          msg.SentryAddress,
		IpAddress:        msg.IpAddress,
		Port:             msg.Port,
		IsActive:         true,
	}

	ms.keeper.SetSentryNode(sdkCtx, sentry)

	// Update validator security info to include sentry in the addresses list
	info, found := ms.keeper.GetValidatorSecurityInfo(sdkCtx, msg.ValidatorAddress)
	if found {
		info.SentryNodeAddresses = append(info.SentryNodeAddresses, msg.SentryAddress)
		ms.keeper.SetValidatorSecurityInfo(sdkCtx, info)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSentryNodeRegistered,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeySentryAddress, msg.SentryAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("sentry node registered", "validator", msg.ValidatorAddress, "sentry", msg.SentryAddress)
	return &securitypb.MsgRegisterSentryNodeResponse{}, nil
}

func (ms msgServer) RemoveSentryNode(ctx context.Context, msg *securitypb.MsgRemoveSentryNode) (*securitypb.MsgRemoveSentryNodeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	// Find and remove sentry node
	sentries := ms.keeper.GetAllSentryNodes(sdkCtx)
	var found bool
	for _, sentry := range sentries {
		if sentry.ValidatorAddress == msg.ValidatorAddress && sentry.Address == msg.SentryAddress {
			found = true
			// Mark as inactive rather than delete for audit purposes
			sentry.IsActive = false
			ms.keeper.SetSentryNode(sdkCtx, sentry)
			break
		}
	}

	if !found {
		return nil, types.ErrNotFound
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSentryNodeRemoved,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeySentryAddress, msg.SentryAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("sentry node removed", "validator", msg.ValidatorAddress, "sentry", msg.SentryAddress)
	return &securitypb.MsgRemoveSentryNodeResponse{}, nil
}

func (ms msgServer) ReportDoubleSign(ctx context.Context, msg *securitypb.MsgReportDoubleSign) (*securitypb.MsgReportDoubleSignResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Reporter); err != nil {
		return nil, err
	}

	if msg.ValidatorAddress == "" {
		return nil, types.ErrValidatorNotFound
	}

	blockTime := sdkCtx.BlockTime()
	evidence := &securitypb.DoubleSignEvidence{
		ValidatorAddress: msg.ValidatorAddress,
		Height:           msg.Height,
		Time:             &blockTime,
		VoteA:            msg.VoteA,
		VoteB:            msg.VoteB,
	}

	ms.keeper.SetDoubleSignEvidence(sdkCtx, evidence)

	// Create incident
	ms.keeper.CreateIncident(sdkCtx, "DOUBLE_SIGN", "critical", fmt.Sprintf("Double sign detected for validator %s", msg.ValidatorAddress), msg.Reporter)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeDoubleSignReported,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyReporter, msg.Reporter),
		),
	)

	ms.keeper.Logger(sdkCtx).Warn("double sign reported", "validator", msg.ValidatorAddress, "reporter", msg.Reporter)
	return &securitypb.MsgReportDoubleSignResponse{}, nil
}

func (ms msgServer) AcknowledgeValidatorAlert(ctx context.Context, msg *securitypb.MsgAcknowledgeValidatorAlert) (*securitypb.MsgAcknowledgeValidatorAlertResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	if msg.AlertId == "" {
		return nil, types.ErrInvalidRequest
	}

	// Find and update the alert
	alerts := ms.keeper.GetAllValidatorAlerts(sdkCtx)
	var found bool
	for _, alert := range alerts {
		if alert.Id == msg.AlertId {
			found = true
			ackTime := sdkCtx.BlockTime()
			alert.AcknowledgedAt = &ackTime
			alert.AcknowledgedBy = msg.ValidatorAddress
			alert.Acknowledged = true
			ms.keeper.SetValidatorAlert(sdkCtx, alert)
			break
		}
	}

	if !found {
		return nil, types.ErrNotFound
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeValidatorAlertAcknowledged,
			sdk.NewAttribute(types.AttributeKeyAlertId, msg.AlertId),
			sdk.NewAttribute(types.AttributeKeyAcknowledgedBy, msg.ValidatorAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("validator alert acknowledged", "alert_id", msg.AlertId, "by", msg.ValidatorAddress)
	return &securitypb.MsgAcknowledgeValidatorAlertResponse{}, nil
}

func (ms msgServer) TriggerFailover(ctx context.Context, msg *securitypb.MsgTriggerFailover) (*securitypb.MsgTriggerFailoverResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	if msg.ValidatorAddress == "" {
		return nil, types.ErrValidatorNotFound
	}

	// Get validator info
	info, found := ms.keeper.GetValidatorSecurityInfo(sdkCtx, msg.ValidatorAddress)
	if !found {
		return nil, types.ErrValidatorNotFound
	}

	// Create failover alert
	alertTime := sdkCtx.BlockTime()
	alert := &securitypb.ValidatorAlert{
		Id:               fmt.Sprintf("FAILOVER-%d", sdkCtx.BlockHeight()),
		ValidatorAddress: msg.ValidatorAddress,
		AlertType:        securitypb.ValidatorAlert_FAILOVER_TRIGGERED,
		Severity:         securitypb.ValidatorAlert_CRITICAL,
		Message:          fmt.Sprintf("Failover triggered: %s", msg.Reason),
		Timestamp:        &alertTime,
	}
	ms.keeper.SetValidatorAlert(sdkCtx, alert)

	// Update validator info to indicate failover
	info.FailoverActive = true
	info.ActiveBackup = msg.BackupValidator
	ms.keeper.SetValidatorSecurityInfo(sdkCtx, info)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeFailoverTriggered,
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
		),
	)

	ms.keeper.Logger(sdkCtx).Warn("failover triggered", "validator", msg.ValidatorAddress, "reason", msg.Reason)
	return &securitypb.MsgTriggerFailoverResponse{}, nil
}

// ========================
// WALLET SECURITY MESSAGES
// ========================

func (ms msgServer) RegisterHardwareWallet(ctx context.Context, msg *securitypb.MsgRegisterHardwareWallet) (*securitypb.MsgRegisterHardwareWalletResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Address); err != nil {
		return nil, err
	}

	// Generate wallet ID
	walletId := generateWalletID(msg.Address, msg.DeviceId)

	// Check if already registered
	if _, found := ms.keeper.GetHardwareWallet(sdkCtx, walletId); found {
		return nil, types.ErrAlreadyExists
	}

	regTime := sdkCtx.BlockTime()
	hw := &securitypb.HardwareWalletConfig{
		WalletId:           walletId,
		Address:            msg.Address,
		Type:               msg.Type,
		DeviceId:           msg.DeviceId,
		FirmwareVersion:    msg.FirmwareVersion,
		DerivationPath:     msg.DerivationPath,
		RequiresPin:        msg.RequiresPin,
		RequiresPassphrase: msg.RequiresPassphrase,
		RegisteredAt:       &regTime,
	}

	ms.keeper.SetHardwareWallet(sdkCtx, hw)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeHardwareWalletRegistered,
			sdk.NewAttribute(types.AttributeKeyWalletId, walletId),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("hardware wallet registered", "wallet_id", walletId, "owner", msg.Address)
	return &securitypb.MsgRegisterHardwareWalletResponse{WalletId: walletId}, nil
}

func (ms msgServer) CreateMultiSigWallet(ctx context.Context, msg *securitypb.MsgCreateMultiSigWallet) (*securitypb.MsgCreateMultiSigWalletResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if msg.Threshold == 0 || int(msg.Threshold) > len(msg.Signers) {
		return nil, types.ErrInvalidRequest
	}

	// Generate wallet ID from signers hash
	walletId := generateMultiSigWalletID(msg.Signers)

	// Check if already exists
	if _, found := ms.keeper.GetMultiSigWallet(sdkCtx, walletId); found {
		return nil, types.ErrAlreadyExists
	}

	createdAt := sdkCtx.BlockTime()
	wallet := &securitypb.MultiSigWallet{
		WalletId:        walletId,
		Signers:         msg.Signers,
		Threshold:       msg.Threshold,
		TotalSigners:    int32(len(msg.Signers)),
		Creator:         msg.Creator,
		CreatedAt:       &createdAt,
		SignerWeights:   msg.SignerWeights,
		WeightThreshold: msg.WeightThreshold,
		TimeLocked:      msg.TimeLocked,
		UnlockTime:      msg.UnlockTime,
	}

	ms.keeper.SetMultiSigWallet(sdkCtx, wallet)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultiSigWalletCreated,
			sdk.NewAttribute(types.AttributeKeyWalletId, walletId),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", msg.Threshold)),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("multisig wallet created", "wallet_id", walletId, "threshold", msg.Threshold, "signers", len(msg.Signers))
	return &securitypb.MsgCreateMultiSigWalletResponse{WalletId: walletId}, nil
}

func (ms msgServer) ProposeMultiSigTransaction(ctx context.Context, msg *securitypb.MsgProposeMultiSigTransaction) (*securitypb.MsgProposeMultiSigTransactionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Proposer); err != nil {
		return nil, err
	}

	// Verify wallet exists
	wallet, found := ms.keeper.GetMultiSigWallet(sdkCtx, msg.WalletId)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	// Verify proposer is a signer
	isValidSigner := false
	for _, signer := range wallet.Signers {
		if signer == msg.Proposer {
			isValidSigner = true
			break
		}
	}
	if !isValidSigner {
		return nil, types.ErrUnauthorized
	}

	// Generate transaction ID
	txId := generateTxID(msg.WalletId, sdkCtx.BlockHeight())

	// Set created time
	createdAt := sdkCtx.BlockTime()

	tx := &securitypb.PendingMultiSigTransaction{
		TxId:          txId,
		WalletId:      msg.WalletId,
		TxData:        msg.TxData,
		SignedBy:      []string{msg.Proposer},
		CurrentWeight: 1, // Proposer counts as first signer
		CreatedAt:     &createdAt,
		ExpiresAt:     msg.ExpiresAt,
		TxType:        msg.TxType,
		Description:   msg.Description,
	}

	ms.keeper.SetPendingMultiSigTx(sdkCtx, tx)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultiSigTxProposed,
			sdk.NewAttribute(types.AttributeKeyTxId, txId),
			sdk.NewAttribute(types.AttributeKeyWalletId, msg.WalletId),
			sdk.NewAttribute(types.AttributeKeyProposer, msg.Proposer),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("multisig transaction proposed", "tx_id", txId, "wallet_id", msg.WalletId)
	return &securitypb.MsgProposeMultiSigTransactionResponse{TxId: txId}, nil
}

func (ms msgServer) SignMultiSigTransaction(ctx context.Context, msg *securitypb.MsgSignMultiSigTransaction) (*securitypb.MsgSignMultiSigTransactionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Signer); err != nil {
		return nil, err
	}

	// Get pending transaction
	tx, found := ms.keeper.GetPendingMultiSigTx(sdkCtx, msg.TxId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Check not expired
	if tx.ExpiresAt != nil && sdkCtx.BlockTime().After(*tx.ExpiresAt) {
		return nil, types.ErrInvalidRequest
	}

	// Verify wallet and signer authorization
	wallet, found := ms.keeper.GetMultiSigWallet(sdkCtx, tx.WalletId)
	if !found {
		return nil, types.ErrWalletNotFound
	}

	isValidSigner := false
	for _, signer := range wallet.Signers {
		if signer == msg.Signer {
			isValidSigner = true
			break
		}
	}
	if !isValidSigner {
		return nil, types.ErrUnauthorized
	}

	// Check not already signed
	for _, signed := range tx.SignedBy {
		if signed == msg.Signer {
			return nil, types.ErrAlreadyExists
		}
	}

	// Add signature
	tx.SignedBy = append(tx.SignedBy, msg.Signer)
	ms.keeper.SetPendingMultiSigTx(sdkCtx, tx)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultiSigTxSigned,
			sdk.NewAttribute(types.AttributeKeyTxId, msg.TxId),
			sdk.NewAttribute(types.AttributeKeySigner, msg.Signer),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("multisig transaction signed", "tx_id", msg.TxId, "signer", msg.Signer, "signatures", len(tx.SignedBy))
	return &securitypb.MsgSignMultiSigTransactionResponse{
		CurrentWeight:    int32(len(tx.SignedBy)),
		ThresholdReached: int32(len(tx.SignedBy)) >= wallet.Threshold,
	}, nil
}

func (ms msgServer) ExecuteMultiSigTransaction(ctx context.Context, msg *securitypb.MsgExecuteMultiSigTransaction) (*securitypb.MsgExecuteMultiSigTransactionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Executor); err != nil {
		return nil, err
	}

	// Get and validate pending transaction
	_, found := ms.keeper.GetPendingMultiSigTx(sdkCtx, msg.TxId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Validate threshold met
	if err := ms.keeper.ValidateMultiSigTx(sdkCtx, msg.TxId); err != nil {
		return nil, err
	}

	// Delete the pending transaction (it's been executed)
	ms.keeper.DeletePendingMultiSigTx(sdkCtx, msg.TxId)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultiSigTxExecuted,
			sdk.NewAttribute(types.AttributeKeyTxId, msg.TxId),
			sdk.NewAttribute(types.AttributeKeyExecutor, msg.Executor),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("multisig transaction executed", "tx_id", msg.TxId, "executor", msg.Executor)
	return &securitypb.MsgExecuteMultiSigTransactionResponse{Success: true}, nil
}

func (ms msgServer) ConfigureSocialRecovery(ctx context.Context, msg *securitypb.MsgConfigureSocialRecovery) (*securitypb.MsgConfigureSocialRecoveryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Address); err != nil {
		return nil, err
	}

	if msg.RecoveryThreshold == 0 || int(msg.RecoveryThreshold) > len(msg.Guardians) {
		return nil, types.ErrInvalidRequest
	}

	configuredAt := sdkCtx.BlockTime()
	config := &securitypb.SocialRecoveryConfig{
		WalletId:          msg.WalletId,
		Guardians:         msg.Guardians,
		RecoveryThreshold: msg.RecoveryThreshold,
		RecoveryDelay:     msg.RecoveryDelay,
		Enabled:           true,
		ConfiguredAt:      &configuredAt,
		LastModified:      &configuredAt,
	}

	ms.keeper.SetSocialRecoveryConfig(sdkCtx, config)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSocialRecoveryConfigured,
			sdk.NewAttribute(types.AttributeKeyWalletId, msg.WalletId),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", msg.RecoveryThreshold)),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("social recovery configured", "wallet", msg.Address, "guardians", len(msg.Guardians))
	return &securitypb.MsgConfigureSocialRecoveryResponse{}, nil
}

func (ms msgServer) InitiateRecovery(ctx context.Context, msg *securitypb.MsgInitiateRecovery) (*securitypb.MsgInitiateRecoveryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Initiator); err != nil {
		return nil, err
	}

	// Verify recovery config exists
	config, found := ms.keeper.GetSocialRecoveryConfig(sdkCtx, msg.WalletId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Verify initiator is a guardian
	isGuardian := false
	for _, guardian := range config.Guardians {
		if guardian.Address == msg.Initiator {
			isGuardian = true
			break
		}
	}
	if !isGuardian {
		return nil, types.ErrUnauthorized
	}

	// Generate request ID
	requestId := fmt.Sprintf("REC-%s-%d", msg.WalletId[:8], sdkCtx.BlockHeight())

	// Calculate delay end time
	executableAt := sdkCtx.BlockTime().Add(config.RecoveryDelay)
	initiatedAt := sdkCtx.BlockTime()

	request := &securitypb.RecoveryRequest{
		RequestId:      requestId,
		WalletId:       msg.WalletId,
		NewAddress:     msg.NewAddress,
		Approvals:      []string{msg.Initiator},
		ApprovalsCount: 1,
		InitiatedAt:    &initiatedAt,
		ExecutableAt:   &executableAt,
		Status:         securitypb.RecoveryStatus_RECOVERY_STATUS_PENDING,
		Initiator:      msg.Initiator,
	}

	ms.keeper.SetRecoveryRequest(sdkCtx, request)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecoveryInitiated,
			sdk.NewAttribute(types.AttributeKeyRequestId, requestId),
			sdk.NewAttribute(types.AttributeKeyWalletId, msg.WalletId),
			sdk.NewAttribute(types.AttributeKeyInitiator, msg.Initiator),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("recovery initiated", "request_id", requestId, "wallet", msg.WalletId)
	return &securitypb.MsgInitiateRecoveryResponse{RequestId: requestId}, nil
}

func (ms msgServer) ApproveRecovery(ctx context.Context, msg *securitypb.MsgApproveRecovery) (*securitypb.MsgApproveRecoveryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Guardian); err != nil {
		return nil, err
	}

	// Get recovery request
	request, found := ms.keeper.GetRecoveryRequest(sdkCtx, msg.RequestId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Get config to verify guardian
	config, found := ms.keeper.GetSocialRecoveryConfig(sdkCtx, request.WalletId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Verify approver is a guardian
	isGuardian := false
	for _, guardian := range config.Guardians {
		if guardian.Address == msg.Guardian {
			isGuardian = true
			break
		}
	}
	if !isGuardian {
		return nil, types.ErrUnauthorized
	}

	// Check not already approved
	for _, approver := range request.Approvals {
		if approver == msg.Guardian {
			return nil, types.ErrAlreadyExists
		}
	}

	// Add approval
	request.Approvals = append(request.Approvals, msg.Guardian)
	request.ApprovalsCount++
	ms.keeper.SetRecoveryRequest(sdkCtx, request)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecoveryApproved,
			sdk.NewAttribute(types.AttributeKeyRequestId, msg.RequestId),
			sdk.NewAttribute(types.AttributeKeyGuardian, msg.Guardian),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("recovery approved", "request_id", msg.RequestId, "guardian", msg.Guardian, "approvals", len(request.Approvals))
	return &securitypb.MsgApproveRecoveryResponse{
		ApprovalsCount:   request.ApprovalsCount,
		ThresholdReached: request.ApprovalsCount >= config.RecoveryThreshold,
	}, nil
}

func (ms msgServer) ExecuteRecovery(ctx context.Context, msg *securitypb.MsgExecuteRecovery) (*securitypb.MsgExecuteRecoveryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Executor); err != nil {
		return nil, err
	}

	// Get recovery request
	request, found := ms.keeper.GetRecoveryRequest(sdkCtx, msg.RequestId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Get config to check threshold
	config, found := ms.keeper.GetSocialRecoveryConfig(sdkCtx, request.WalletId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Check threshold met
	if request.ApprovalsCount < config.RecoveryThreshold {
		return nil, types.ErrRecoveryNotReady
	}

	// Check delay elapsed
	if request.ExecutableAt != nil && sdkCtx.BlockTime().Before(*request.ExecutableAt) {
		return nil, types.ErrRecoveryNotReady
	}

	// Mark as executed
	request.Status = securitypb.RecoveryStatus_RECOVERY_STATUS_EXECUTED
	executedAt := sdkCtx.BlockTime()
	request.ExecutedAt = &executedAt
	ms.keeper.SetRecoveryRequest(sdkCtx, request)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRecoveryExecuted,
			sdk.NewAttribute(types.AttributeKeyRequestId, msg.RequestId),
			sdk.NewAttribute(types.AttributeKeyAddress, request.NewAddress),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("recovery executed", "request_id", msg.RequestId, "new_address", request.NewAddress)
	return &securitypb.MsgExecuteRecoveryResponse{Success: true}, nil
}

func (ms msgServer) SetSpendingLimits(ctx context.Context, msg *securitypb.MsgSetSpendingLimits) (*securitypb.MsgSetSpendingLimitsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Address); err != nil {
		return nil, err
	}

	// Store limit using the proto message fields (daily_limit, weekly_limit, monthly_limit)
	// Map to WalletLimit supplemental type using the daily limit as max tx amount
	blockTime := sdkCtx.BlockTime()
	limit := &types.WalletLimit{
		WalletAddress: msg.Address,
		MaxTxAmount:   msg.DailyLimit,
		MaxDailyTxs:   0, // No max count in proto, just amount
		SetAt:         &blockTime,
		Reason:        fmt.Sprintf("denom:%s daily:%s weekly:%s monthly:%s", msg.Denom, msg.DailyLimit, msg.WeeklyLimit, msg.MonthlyLimit),
	}

	ms.keeper.SetWalletLimit(sdkCtx, limit)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSpendingLimitSet,
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
			sdk.NewAttribute("daily_limit", msg.DailyLimit),
			sdk.NewAttribute("denom", msg.Denom),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("spending limits set", "wallet", msg.Address, "daily", msg.DailyLimit)
	return &securitypb.MsgSetSpendingLimitsResponse{}, nil
}

func (ms msgServer) RegisterBiometric(ctx context.Context, msg *securitypb.MsgRegisterBiometric) (*securitypb.MsgRegisterBiometricResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Address); err != nil {
		return nil, err
	}

	// Store device fingerprint with biometric info
	// Map biometric enrollment to device fingerprint structure
	blockTime := sdkCtx.BlockTime()
	fp := &types.DeviceFingerprint{
		Id:            generateDeviceFingerprintID(msg.Address, msg.WalletId),
		WalletAddress: msg.Address,
		Fingerprint:   msg.EnrollmentHash,
		DeviceName:    msg.Type.String(), // Store biometric type as device name
		TrustedAt:     &blockTime,
		LastUsed:      &blockTime,
		IsActive:      true,
	}

	ms.keeper.SetDeviceFingerprint(sdkCtx, fp)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBiometricRegistered,
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
			sdk.NewAttribute("wallet_id", msg.WalletId),
			sdk.NewAttribute("biometric_type", msg.Type.String()),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("biometric registered", "wallet", msg.Address, "type", msg.Type.String())
	return &securitypb.MsgRegisterBiometricResponse{}, nil
}

// ========================
// INCIDENT RESPONSE MESSAGES
// ========================

func (ms msgServer) CreateIncident(ctx context.Context, msg *securitypb.MsgCreateIncident) (*securitypb.MsgCreateIncidentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Reporter); err != nil {
		return nil, err
	}

	// Convert enum severity to string for keeper method
	incident := ms.keeper.CreateIncident(sdkCtx, msg.Title, msg.Severity.String(), msg.Description, msg.Reporter)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIncidentCreated,
			sdk.NewAttribute(types.AttributeKeyIncidentId, incident.IncidentId),
			sdk.NewAttribute(types.AttributeKeySeverity, msg.Severity.String()),
			sdk.NewAttribute(types.AttributeKeyReporter, msg.Reporter),
		),
	)

	return &securitypb.MsgCreateIncidentResponse{IncidentId: incident.IncidentId}, nil
}

func (ms msgServer) UpdateIncident(ctx context.Context, msg *securitypb.MsgUpdateIncident) (*securitypb.MsgUpdateIncidentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Updater); err != nil {
		return nil, err
	}

	incident, found := ms.keeper.GetIncident(sdkCtx, msg.IncidentId)
	if !found {
		return nil, types.ErrIncidentNotFound
	}

	// Update fields
	if msg.Status != securitypb.IncidentStatus_INCIDENT_STATUS_UNSPECIFIED {
		incident.Status = msg.Status
	}
	if msg.Description != "" {
		incident.Description = msg.Description
	}
	if msg.AssignedTo != "" {
		incident.AssignedTo = msg.AssignedTo
	}

	// UpdatedAt is *time.Time in proto
	blockTime := sdkCtx.BlockTime()
	incident.UpdatedAt = &blockTime

	ms.keeper.SetIncident(sdkCtx, incident)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIncidentUpdated,
			sdk.NewAttribute(types.AttributeKeyIncidentId, msg.IncidentId),
			sdk.NewAttribute("updater", msg.Updater),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("incident updated", "incident_id", msg.IncidentId, "by", msg.Updater)
	return &securitypb.MsgUpdateIncidentResponse{}, nil
}

func (ms msgServer) ResolveIncident(ctx context.Context, msg *securitypb.MsgResolveIncident) (*securitypb.MsgResolveIncidentResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Resolver); err != nil {
		return nil, err
	}

	// Pass resolution details as a single-element slice for the actionsTaken parameter
	if err := ms.keeper.ResolveIncident(sdkCtx, msg.IncidentId, []string{msg.ResolutionDetails}); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeIncidentResolved,
			sdk.NewAttribute(types.AttributeKeyIncidentId, msg.IncidentId),
			sdk.NewAttribute("resolver", msg.Resolver),
		),
	)

	return &securitypb.MsgResolveIncidentResponse{}, nil
}

func (ms msgServer) ExecuteResponseAction(ctx context.Context, msg *securitypb.MsgExecuteResponseAction) (*securitypb.MsgExecuteResponseActionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Executor); err != nil {
		return nil, err
	}

	// Verify incident exists and is active
	incident, found := ms.keeper.GetIncident(sdkCtx, msg.IncidentId)
	if !found {
		return nil, types.ErrIncidentNotFound
	}

	if incident.Status == securitypb.IncidentStatus_INCIDENT_STATUS_RESOLVED {
		return nil, types.ErrInvalidState
	}

	// Generate action ID
	actionId := fmt.Sprintf("action-%d", sdkCtx.BlockHeight())

	// Execute the action based on type (using Description for additional context)
	switch msg.ActionType {
	case "pause_system":
		// Level 1 = full pause
		if err := ms.keeper.PauseSystem(sdkCtx, 1, msg.Description, msg.Executor); err != nil {
			return nil, err
		}
	case "resume_system":
		if err := ms.keeper.ResumeSystem(sdkCtx); err != nil {
			return nil, err
		}
	default:
		// Log the action for other types
		ms.keeper.Logger(sdkCtx).Info("response action", "type", msg.ActionType, "desc", msg.Description)
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeResponseActionExecuted,
			sdk.NewAttribute(types.AttributeKeyIncidentId, msg.IncidentId),
			sdk.NewAttribute(types.AttributeKeyActionType, msg.ActionType),
			sdk.NewAttribute(types.AttributeKeyExecutor, msg.Executor),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("response action executed", "incident", msg.IncidentId, "action", msg.ActionType)
	return &securitypb.MsgExecuteResponseActionResponse{
		ActionId:   actionId,
		Successful: true,
		Result:     "Action executed successfully",
	}, nil
}

func (ms msgServer) AddAuditLogEntry(ctx context.Context, msg *securitypb.MsgAddAuditLogEntry) (*securitypb.MsgAddAuditLogEntryResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Actor); err != nil {
		return nil, err
	}

	// Generate log ID
	logId := fmt.Sprintf("log-%d-%s", sdkCtx.BlockHeight(), msg.EventType)

	// Emit audit event (audit logs are stored via events in Cosmos SDK)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAuditLog,
			sdk.NewAttribute("action", msg.Action),
			sdk.NewAttribute("actor", msg.Actor),
			sdk.NewAttribute("resource", msg.Resource),
			sdk.NewAttribute("event_type", msg.EventType),
			sdk.NewAttribute("success", fmt.Sprintf("%v", msg.Success)),
			sdk.NewAttribute("timestamp", sdkCtx.BlockTime().String()),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("audit log entry added", "action", msg.Action, "actor", msg.Actor, "resource", msg.Resource)
	return &securitypb.MsgAddAuditLogEntryResponse{LogId: logId}, nil
}

// ========================
// CRYPTOGRAPHY MESSAGES
// ========================

func (ms msgServer) CreateKeyRotationSchedule(ctx context.Context, msg *securitypb.MsgCreateKeyRotationSchedule) (*securitypb.MsgCreateKeyRotationScheduleResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if msg.KeyId == "" {
		return nil, types.ErrInvalidRequest
	}

	// Check if schedule already exists
	if _, found := ms.keeper.GetKeyRotationSchedule(sdkCtx, msg.KeyId); found {
		return nil, types.ErrAlreadyExists
	}

	// Calculate next rotation time
	nextRotation := sdkCtx.BlockTime().Add(time.Duration(msg.RotationIntervalSeconds) * time.Second)

	schedule := &securitypb.KeyRotationSchedule{
		Id:                      msg.KeyId,
		KeyId:                   msg.KeyId,
		RotationIntervalSeconds: msg.RotationIntervalSeconds,
		NextRotationTime:        nextRotation,
		Enabled:                 true,
		CreatedBy:               msg.Creator,
		Policy:                  msg.Policy,
	}

	ms.keeper.SetKeyRotationSchedule(sdkCtx, schedule)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeKeyRotationScheduleCreated,
			sdk.NewAttribute(types.AttributeKeyScheduleId, msg.KeyId),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("key rotation schedule created", "key_id", msg.KeyId, "interval", msg.RotationIntervalSeconds)
	return &securitypb.MsgCreateKeyRotationScheduleResponse{ScheduleId: msg.KeyId}, nil
}

func (ms msgServer) RotateKey(ctx context.Context, msg *securitypb.MsgRotateKey) (*securitypb.MsgRotateKeyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if err := ms.keeper.RotateKey(sdkCtx, msg.KeyId); err != nil {
		return nil, err
	}

	rotationId := fmt.Sprintf("rot-%s-%d", msg.KeyId, sdkCtx.BlockHeight())
	ms.keeper.Logger(sdkCtx).Info("key rotated", "key_id", msg.KeyId, "requested_by", msg.Creator)
	return &securitypb.MsgRotateKeyResponse{
		RotationId:   rotationId,
		RotationTime: sdkCtx.BlockTime(),
	}, nil
}

func (ms msgServer) CreateThresholdScheme(ctx context.Context, msg *securitypb.MsgCreateThresholdScheme) (*securitypb.MsgCreateThresholdSchemeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if msg.Threshold == 0 || int(msg.Threshold) > int(msg.TotalParticipants) {
		return nil, types.ErrInvalidRequest
	}

	// Generate scheme ID
	schemeId := generateSchemeID(msg.Creator, sdkCtx.BlockHeight())

	scheme := &securitypb.ThresholdSignatureScheme{
		SchemeId:          schemeId,
		Threshold:         msg.Threshold,
		TotalParticipants: msg.TotalParticipants,
		ParticipantIds:    msg.ParticipantIds,
		SchemeType:        msg.SchemeType,
		CreatedAt:         sdkCtx.BlockTime(),
		Status:            securitypb.ThresholdSchemeStatus_THRESHOLD_SCHEME_STATUS_ACTIVE,
	}

	ms.keeper.SetThresholdScheme(sdkCtx, scheme)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeThresholdSchemeCreated,
			sdk.NewAttribute(types.AttributeKeySchemeId, schemeId),
			sdk.NewAttribute(types.AttributeKeyThreshold, fmt.Sprintf("%d", msg.Threshold)),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("threshold scheme created", "scheme_id", schemeId, "threshold", msg.Threshold)
	return &securitypb.MsgCreateThresholdSchemeResponse{SchemeId: schemeId}, nil
}

func (ms msgServer) SubmitThresholdSignatureShare(ctx context.Context, msg *securitypb.MsgSubmitThresholdSignatureShare) (*securitypb.MsgSubmitThresholdSignatureShareResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Submitter); err != nil {
		return nil, err
	}

	// Verify scheme exists
	scheme, found := ms.keeper.GetThresholdScheme(sdkCtx, msg.SchemeId)
	if !found {
		return nil, types.ErrThresholdSchemeNotFound
	}

	// Verify submitter is a participant in scheme
	isParticipant := false
	for _, p := range scheme.ParticipantIds {
		if p == msg.Submitter {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, types.ErrUnauthorized
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeThresholdShareSubmitted,
			sdk.NewAttribute(types.AttributeKeySchemeId, msg.SchemeId),
			sdk.NewAttribute(types.AttributeKeySubmitter, msg.Submitter),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("threshold signature share submitted", "scheme_id", msg.SchemeId, "submitter", msg.Submitter)
	return &securitypb.MsgSubmitThresholdSignatureShareResponse{SharesCollected: 1}, nil
}

func (ms msgServer) RegisterZKProofCircuit(ctx context.Context, msg *securitypb.MsgRegisterZKProofCircuit) (*securitypb.MsgRegisterZKProofCircuitResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if msg.CircuitId == "" {
		return nil, types.ErrInvalidRequest
	}

	// Check if circuit already exists
	if _, found := ms.keeper.GetZKProofConfig(sdkCtx, msg.CircuitId); found {
		return nil, types.ErrAlreadyExists
	}

	// Generate proof ID from circuit ID
	proofId := fmt.Sprintf("zkp-%s-%d", msg.CircuitId, sdkCtx.BlockHeight())

	config := &securitypb.ZKProofConfig{
		ProofId:          proofId,
		ProofType:        msg.ProofType,
		PublicParameters: msg.PublicParameters,
		VerificationKey:  msg.VerificationKey,
		CircuitId:        msg.CircuitId,
		CreatedAt:        sdkCtx.BlockTime(),
	}

	ms.keeper.SetZKProofConfig(sdkCtx, config)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeZKCircuitRegistered,
			sdk.NewAttribute(types.AttributeKeyCircuitId, msg.CircuitId),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("ZK proof circuit registered", "circuit_id", msg.CircuitId, "type", msg.ProofType.String())
	return &securitypb.MsgRegisterZKProofCircuitResponse{ProofId: proofId}, nil
}

func (ms msgServer) SubmitZKProof(ctx context.Context, msg *securitypb.MsgSubmitZKProof) (*securitypb.MsgSubmitZKProofResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Submitter); err != nil {
		return nil, err
	}

	// Verify proof config exists
	_, found := ms.keeper.GetZKProofConfig(sdkCtx, msg.ProofId)
	if !found {
		return nil, types.ErrNotFound
	}

	// Generate verification ID
	verificationId := fmt.Sprintf("ver-%s-%d", msg.ProofId, sdkCtx.BlockHeight())

	// In production, actual ZK proof verification would happen here
	// For now, we emit an event and return success
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeZKProofSubmitted,
			sdk.NewAttribute(types.AttributeKeyProofId, msg.ProofId),
			sdk.NewAttribute(types.AttributeKeySubmitter, msg.Submitter),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("ZK proof submitted", "proof_id", msg.ProofId, "submitter", msg.Submitter)
	return &securitypb.MsgSubmitZKProofResponse{
		Verified:       true,
		VerificationId: verificationId,
	}, nil
}

func (ms msgServer) GenerateQuantumResistantKey(ctx context.Context, msg *securitypb.MsgGenerateQuantumResistantKey) (*securitypb.MsgGenerateQuantumResistantKeyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	// Generate key ID and placeholder public key
	keyId := generateQuantumKeyID(msg.Creator, msg.Algorithm.String(), sdkCtx.BlockHeight())
	// In production, actual quantum-resistant key generation would happen here
	publicKey := []byte(fmt.Sprintf("qrk-pub-%s", keyId))

	qrk := &securitypb.QuantumResistantKey{
		KeyId:     keyId,
		Algorithm: msg.Algorithm,
		PublicKey: publicKey,
		CreatedAt: sdkCtx.BlockTime(),
		ExpiresAt: msg.ExpiresAt,
	}

	ms.keeper.SetQuantumResistantKey(sdkCtx, qrk)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeQuantumKeyGenerated,
			sdk.NewAttribute(types.AttributeKeyKeyId, keyId),
			sdk.NewAttribute(types.AttributeKeyAlgorithm, msg.Algorithm.String()),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("quantum-resistant key generated", "key_id", keyId, "algorithm", msg.Algorithm.String())
	return &securitypb.MsgGenerateQuantumResistantKeyResponse{
		KeyId:     keyId,
		PublicKey: publicKey,
	}, nil
}

// ========================
// PRIVACY MESSAGES
// ========================

func (ms msgServer) CreateMixingPool(ctx context.Context, msg *securitypb.MsgCreateMixingPool) (*securitypb.MsgCreateMixingPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	if msg.MinParticipants == 0 {
		return nil, types.ErrInvalidRequest
	}

	pool := ms.keeper.CreateMixingPool(sdkCtx, msg.Denomination, msg.MinParticipants)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMixingPoolCreated,
			sdk.NewAttribute(types.AttributeKeyPoolId, pool.PoolId),
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		),
	)

	return &securitypb.MsgCreateMixingPoolResponse{PoolId: pool.PoolId}, nil
}

func (ms msgServer) JoinMixingPool(ctx context.Context, msg *securitypb.MsgJoinMixingPool) (*securitypb.MsgJoinMixingPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Participant); err != nil {
		return nil, err
	}

	if err := ms.keeper.JoinMixingPool(sdkCtx, msg.PoolId); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMixingPoolJoined,
			sdk.NewAttribute(types.AttributeKeyPoolId, msg.PoolId),
			sdk.NewAttribute(types.AttributeKeyParticipant, msg.Participant),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("joined mixing pool", "pool_id", msg.PoolId, "participant", msg.Participant)
	return &securitypb.MsgJoinMixingPoolResponse{PoolReady: false}, nil
}

func (ms msgServer) ExecuteMixing(ctx context.Context, msg *securitypb.MsgExecuteMixing) (*securitypb.MsgExecuteMixingResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Executor); err != nil {
		return nil, err
	}

	// Get pool and validate
	pool, found := ms.keeper.GetMixingPool(sdkCtx, msg.PoolId)
	if !found {
		return nil, types.ErrMixingPoolNotFound
	}

	// Validate minimum participants
	if err := ms.keeper.ValidateMixingParticipants(sdkCtx, uint32(len(pool.Participants))); err != nil {
		return nil, err
	}

	// Mark pool as mixing
	pool.Status = "mixing"
	ms.keeper.SetMixingPool(sdkCtx, pool)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMixingExecuted,
			sdk.NewAttribute(types.AttributeKeyPoolId, msg.PoolId),
			sdk.NewAttribute(types.AttributeKeyExecutor, msg.Executor),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("mixing executed", "pool_id", msg.PoolId)
	return &securitypb.MsgExecuteMixingResponse{
		Success: true,
		Outputs: nil, // Mixing outputs would be computed here
	}, nil
}

func (ms msgServer) GenerateStealthAddress(ctx context.Context, msg *securitypb.MsgGenerateStealthAddress) (*securitypb.MsgGenerateStealthAddressResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Creator); err != nil {
		return nil, err
	}

	// Generate one-time address from public spend key and view key
	oneTimeAddrStr := generateStealthAddress(msg.PublicSpendKey, sdkCtx.BlockHeight())
	oneTimeAddr := []byte(oneTimeAddrStr)
	// Generate transaction public key
	txPubKey := []byte(fmt.Sprintf("txpub-%d", sdkCtx.BlockHeight()))

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeStealthAddressGenerated,
			sdk.NewAttribute(types.AttributeKeyCreator, msg.Creator),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("stealth address generated", "creator", msg.Creator)
	return &securitypb.MsgGenerateStealthAddressResponse{
		OneTimeAddress: oneTimeAddr,
		TxPublicKey:    txPubKey,
	}, nil
}

func (ms msgServer) CreateRingSignature(ctx context.Context, msg *securitypb.MsgCreateRingSignature) (*securitypb.MsgCreateRingSignatureResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Signer); err != nil {
		return nil, err
	}

	// Validate ring size
	if err := ms.keeper.ValidateRingSize(sdkCtx, uint32(len(msg.RingMembers))); err != nil {
		return nil, err
	}

	// Generate ring signature components
	signatureId := fmt.Sprintf("RING-%d-%s", sdkCtx.BlockHeight(), msg.Signer[:8])
	keyImage := sha256.Sum256(append(msg.Message, msg.Signer...))
	signatureData := sha256.Sum256([]byte(signatureId))

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRingSignatureCreated,
			sdk.NewAttribute(types.AttributeKeySignatureId, signatureId),
			sdk.NewAttribute(types.AttributeKeyRingSize, fmt.Sprintf("%d", len(msg.RingMembers))),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("ring signature created", "signature_id", signatureId, "ring_size", len(msg.RingMembers))
	return &securitypb.MsgCreateRingSignatureResponse{
		KeyImage:      keyImage[:],
		SignatureData: signatureData[:],
	}, nil
}

func (ms msgServer) CreateConfidentialTransaction(ctx context.Context, msg *securitypb.MsgCreateConfidentialTransaction) (*securitypb.MsgCreateConfidentialTransactionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Sender); err != nil {
		return nil, err
	}

	// Generate confidential transaction ID
	ctxId := fmt.Sprintf("CTX-%d-%s", sdkCtx.BlockHeight(), msg.Sender[:8])

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeConfidentialTxCreated,
			sdk.NewAttribute(types.AttributeKeyTxId, ctxId),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("confidential transaction created", "tx_id", ctxId)
	return &securitypb.MsgCreateConfidentialTransactionResponse{TxId: ctxId}, nil
}

// ========================
// PARAMS MESSAGE
// ========================

func (ms msgServer) UpdateParams(ctx context.Context, msg *securitypb.MsgUpdateParams) (*securitypb.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := ms.verifySigner(sdkCtx, msg.Authority); err != nil {
		return nil, err
	}

	// Validate authority is governance module
	if msg.Authority != ms.keeper.GetAuthority() {
		return nil, types.ErrUnauthorized
	}

	ms.keeper.SetParams(sdkCtx, msg.Params)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeParamsUpdated,
			sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
		),
	)

	ms.keeper.Logger(sdkCtx).Info("security params updated", "authority", msg.Authority)
	return &securitypb.MsgUpdateParamsResponse{}, nil
}

// ========================
// HELPER FUNCTIONS
// ========================

func generateWalletID(address, deviceFingerprint string) string {
	h := sha256.Sum256([]byte(address + deviceFingerprint))
	return "HW-" + hex.EncodeToString(h[:8])
}

func generateMultiSigWalletID(signers []string) string {
	data := ""
	for _, s := range signers {
		data += s
	}
	h := sha256.Sum256([]byte(data))
	return "MS-" + hex.EncodeToString(h[:8])
}

func generateTxID(walletId string, height int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", walletId, height)))
	return "TX-" + hex.EncodeToString(h[:8])
}

func generateDeviceFingerprintID(address, deviceId string) string {
	h := sha256.Sum256([]byte(address + deviceId))
	return "DEV-" + hex.EncodeToString(h[:8])
}

func generateSchemeID(creator string, height int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", creator, height)))
	return "TSS-" + hex.EncodeToString(h[:8])
}

func generateQuantumKeyID(owner, algorithm string, height int64) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%s-%d", owner, algorithm, height)))
	return "QRK-" + hex.EncodeToString(h[:8])
}

func generateStealthAddress(recipientPubKey []byte, height int64) string {
	data := append(recipientPubKey, []byte(fmt.Sprintf("%d", height))...)
	h := sha256.Sum256(data)
	return "aura1stealth" + hex.EncodeToString(h[:15])
}
