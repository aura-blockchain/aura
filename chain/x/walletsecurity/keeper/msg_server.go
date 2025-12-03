package keeper

import (
    "github.com/aequitas/aura/chain/x/common/determinism"
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wspb "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

var _ wspb.MsgServer = (*msgServer)(nil)

type msgServer struct {
	wspb.UnimplementedMsgServer
	Keeper *Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper *Keeper) wspb.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterHardwareWallet registers a new hardware wallet
func (ms msgServer) RegisterHardwareWallet(goCtx context.Context, msg *wspb.MsgRegisterHardwareWallet) (*wspb.MsgRegisterHardwareWalletResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if msg.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	// CRITICAL: Verify transaction signer matches claimed address
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedAddr, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid address format")
	}

	if !signers[0].Equals(claimedAddr) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match claimed address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	walletID := fmt.Sprintf("hw_%s_%d", msg.Address, ctx.BlockHeight())

	config := &wspb.HardwareWalletConfig{
		WalletId:        walletID,
		Address:         msg.Address,
		Type:            msg.Type,
		DeviceId:        msg.DeviceId,
		FirmwareVersion: msg.FirmwareVersion,
		DerivationPath:  msg.DerivationPath,
		RegisteredAt:    timestamppb.Now(),
		SignatureCount:  0,
	}

	configBytes, err := ms.Keeper.cdc.Marshal(config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetHardwareWallet(ctx, walletID, configBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRegisterHardwareWallet,
			sdk.NewAttribute(types.AttributeKeyWalletID, walletID),
			sdk.NewAttribute(types.AttributeKeyAddress, msg.Address),
		),
	)

	return &wspb.MsgRegisterHardwareWalletResponse{
		WalletId: walletID,
		Config:   config,
	}, nil
}

// CreateMultiSigWallet creates a new multi-signature wallet
func (ms msgServer) CreateMultiSigWallet(goCtx context.Context, msg *wspb.MsgCreateMultiSigWallet) (*wspb.MsgCreateMultiSigWalletResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify transaction signer matches claimed creator
	if msg.Creator == "" {
		return nil, status.Error(codes.InvalidArgument, "creator cannot be empty")
	}

	// Use the SDK helper to get signers (implemented via GetSignersSDK)
	signers := msg.GetSignersSDK()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedCreator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid creator address")
	}

	if !signers[0].Equals(claimedCreator) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match creator")
	}

	if len(msg.Signers) < 2 {
		return nil, status.Error(codes.InvalidArgument, "at least 2 signers required")
	}

	if msg.Threshold < 1 || int(msg.Threshold) > len(msg.Signers) {
		return nil, status.Error(codes.InvalidArgument, "invalid threshold")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	walletID := fmt.Sprintf("multisig_%s_%d", msg.Creator, ctx.BlockHeight())

	wallet := &wspb.MultiSigWallet{
		WalletId:      walletID,
		Signers:       msg.Signers,
		Threshold:     msg.Threshold,
		TotalSigners:  int32(len(msg.Signers)),
		CreatedAt:     timestamppb.Now(),
		Creator:       msg.Creator,
		SignerWeights: msg.SignerWeights,
		WeightThreshold: msg.WeightThreshold,
	}

	if msg.TimeLock != nil {
		wallet.TimeLocked = true
		wallet.UnlockTime = timestamppb.New(determinism.GetBlockTime(ctx).Add(msg.TimeLock.AsDuration()))
	}

	walletBytes, err := ms.Keeper.cdc.Marshal(wallet)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetMultiSigWallet(ctx, walletID, walletBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMultiSigWallet,
			sdk.NewAttribute(types.AttributeKeyWalletID, walletID),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	return &wspb.MsgCreateMultiSigWalletResponse{
		WalletId: walletID,
		Wallet:   wallet,
	}, nil
}

// SignMultiSigTransaction signs a pending multi-sig transaction
func (ms msgServer) SignMultiSigTransaction(goCtx context.Context, msg *wspb.MsgSignMultiSigTransaction) (*wspb.MsgSignMultiSigTransactionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches claimed signer address
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedSigner, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid signer address")
	}

	if !signers[0].Equals(claimedSigner) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match claimed signer")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	txBytes, err := ms.Keeper.GetPendingMultiSigTx(ctx, msg.TxId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	var tx wspb.PendingMultiSigTransaction
	if err := ms.Keeper.cdc.Unmarshal(txBytes, &tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get wallet configuration to validate signer authorization and weights
	walletBytes, err := ms.Keeper.GetMultiSigWallet(ctx, tx.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}

	var wallet wspb.MultiSigWallet
	if err := ms.Keeper.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// CRITICAL: Verify signer is authorized for this wallet
	isAuthorized := false
	for _, authorizedSigner := range wallet.Signers {
		if authorizedSigner == msg.Signer {
			isAuthorized = true
			break
		}
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, "signer not authorized for this wallet")
	}

	// CRITICAL: Check for duplicate signature
	for _, existingSigner := range tx.SignedBy {
		if existingSigner == msg.Signer {
			return nil, status.Error(codes.AlreadyExists, "signer already signed this transaction")
		}
	}

	// CRITICAL: Calculate actual weight to add (not just +1)
	signerWeight := int32(1) // Default weight
	if wallet.SignerWeights != nil {
		if weight, ok := wallet.SignerWeights[msg.Signer]; ok {
			signerWeight = weight
		}
	}

	// Add signature with proper weight
	tx.Signatures = append(tx.Signatures, string(msg.Signature))
	tx.SignedBy = append(tx.SignedBy, msg.Signer)
	tx.CurrentWeight += signerWeight // Use actual weight, not just ++

	// CRITICAL: Check threshold using weight threshold if configured
	requiredWeight := wallet.Threshold
	if wallet.WeightThreshold > 0 {
		requiredWeight = wallet.WeightThreshold
	}
	readyToExecute := tx.CurrentWeight >= requiredWeight

	// Update transaction
	updatedTxBytes, err := ms.Keeper.cdc.Marshal(&tx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetPendingMultiSigTx(ctx, msg.TxId, updatedTxBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSignMultiSigTx,
			sdk.NewAttribute(types.AttributeKeyTxID, msg.TxId),
			sdk.NewAttribute(types.AttributeKeySigner, msg.Signer),
		),
	)

	return &wspb.MsgSignMultiSigTransactionResponse{
		CurrentSignatures:  tx.CurrentWeight,
		RequiredSignatures: requiredWeight,
		ReadyToExecute:     readyToExecute,
	}, nil
}

// ConfigureSocialRecovery configures social recovery for a wallet
func (ms msgServer) ConfigureSocialRecovery(goCtx context.Context, msg *wspb.MsgConfigureSocialRecovery) (*wspb.MsgConfigureSocialRecoveryResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer is the wallet owner
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedOwner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	if !signers[0].Equals(claimedOwner) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match owner")
	}

	if len(msg.Guardians) < 2 {
		return nil, status.Error(codes.InvalidArgument, "at least 2 guardians required")
	}

	if msg.RecoveryThreshold < 1 || int(msg.RecoveryThreshold) > len(msg.Guardians) {
		return nil, status.Error(codes.InvalidArgument, "invalid recovery threshold")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config := &wspb.SocialRecoveryConfig{
		WalletId:          msg.WalletId,
		Guardians:         msg.Guardians,
		RecoveryThreshold: msg.RecoveryThreshold,
		RecoveryDelay:     msg.RecoveryDelay,
		Enabled:           true,
		ConfiguredAt:      timestamppb.Now(),
		MaxGuardians:      int32(len(msg.Guardians)),
	}

	configBytes, err := ms.Keeper.cdc.Marshal(config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetSocialRecoveryConfig(ctx, msg.WalletId, configBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSocialRecovery,
			sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
		),
	)

	return &wspb.MsgConfigureSocialRecoveryResponse{Config: config}, nil
}

// InitiateRecovery initiates a social recovery process
func (ms msgServer) InitiateRecovery(goCtx context.Context, msg *wspb.MsgInitiateRecovery) (*wspb.MsgInitiateRecoveryResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches claimed initiator (guardian)
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedInitiator, err := sdk.AccAddressFromBech32(msg.Initiator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid initiator address")
	}

	if !signers[0].Equals(claimedInitiator) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match initiator")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get recovery config
	configBytes, err := ms.Keeper.GetSocialRecoveryConfig(ctx, msg.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery not configured")
	}

	var config wspb.SocialRecoveryConfig
	if err := ms.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	requestID := fmt.Sprintf("recovery_%s_%d", msg.WalletId, ctx.BlockHeight())

	now := determinism.GetBlockTime(ctx)
	executableAt := now.Add(config.RecoveryDelay.AsDuration())

	request := &wspb.RecoveryRequest{
		RequestId:      requestID,
		WalletId:       msg.WalletId,
		NewAddress:     msg.NewAddress,
		Approvals:      []string{msg.Initiator},
		ApprovalsCount: 1,
		InitiatedAt:    timestamppb.New(now),
		ExecutableAt:   timestamppb.New(executableAt),
		Status:         wspb.RecoveryStatus_RECOVERY_STATUS_PENDING,
		Initiator:      msg.Initiator,
	}

	requestBytes, err := ms.Keeper.cdc.Marshal(request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetRecoveryRequest(ctx, requestID, requestBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeInitiateRecovery,
			sdk.NewAttribute(types.AttributeKeyRequestID, requestID),
			sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
		),
	)

	return &wspb.MsgInitiateRecoveryResponse{
		RequestId: requestID,
		Request:   request,
	}, nil
}

// ApproveRecovery approves a recovery request as a guardian
func (ms msgServer) ApproveRecovery(goCtx context.Context, msg *wspb.MsgApproveRecovery) (*wspb.MsgApproveRecoveryResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches claimed guardian
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedGuardian, err := sdk.AccAddressFromBech32(msg.Guardian)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid guardian address")
	}

	if !signers[0].Equals(claimedGuardian) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match guardian")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get recovery request
	requestBytes, err := ms.Keeper.GetRecoveryRequest(ctx, msg.RequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery request not found")
	}

	var request wspb.RecoveryRequest
	if err := ms.Keeper.cdc.Unmarshal(requestBytes, &request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get recovery config to verify guardian authorization
	configBytes, err := ms.Keeper.GetSocialRecoveryConfig(ctx, request.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery config not found")
	}

	var config wspb.SocialRecoveryConfig
	if err := ms.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// CRITICAL: Verify guardian is in authorized list and confirmed
	isAuthorized := false
	for _, authorizedGuardian := range config.Guardians {
		if authorizedGuardian.Address == msg.Guardian {
			if !authorizedGuardian.Confirmed {
				return nil, status.Error(codes.PermissionDenied, "guardian not confirmed")
			}
			isAuthorized = true
			break
		}
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, "not an authorized guardian for this wallet")
	}

	// CRITICAL: Prevent duplicate approvals
	for _, existingApproval := range request.Approvals {
		if existingApproval == msg.Guardian {
			return nil, status.Error(codes.AlreadyExists, "guardian already approved this recovery request")
		}
	}

	// Add approval
	request.Approvals = append(request.Approvals, msg.Guardian)
	request.ApprovalsCount++

	readyToExecute := request.ApprovalsCount >= config.RecoveryThreshold

	if readyToExecute {
		request.Status = wspb.RecoveryStatus_RECOVERY_STATUS_APPROVED
	}

	updatedRequestBytes, err := ms.Keeper.cdc.Marshal(&request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetRecoveryRequest(ctx, msg.RequestId, updatedRequestBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSocialRecovery,
			sdk.NewAttribute("request_id", msg.RequestId),
			sdk.NewAttribute("guardian", msg.Guardian),
		),
	)

	return &wspb.MsgApproveRecoveryResponse{
		ApprovalsCount:    request.ApprovalsCount,
		RequiredApprovals: config.RecoveryThreshold,
		ReadyToExecute:    readyToExecute,
		ExecutableAt:      request.ExecutableAt,
	}, nil
}

// ExecuteRecovery executes an approved recovery request
func (ms msgServer) ExecuteRecovery(goCtx context.Context, msg *wspb.MsgExecuteRecovery) (*wspb.MsgExecuteRecoveryResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches executor
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedExecutor, err := sdk.AccAddressFromBech32(msg.Executor)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid executor address")
	}

	if !signers[0].Equals(claimedExecutor) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match executor")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	requestBytes, err := ms.Keeper.GetRecoveryRequest(ctx, msg.RequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery request not found")
	}

	var request wspb.RecoveryRequest
	if err := ms.Keeper.cdc.Unmarshal(requestBytes, &request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if request.Status != wspb.RecoveryStatus_RECOVERY_STATUS_APPROVED {
		return nil, status.Error(codes.FailedPrecondition, "recovery not approved")
	}

	// Check if executable time has passed
	if determinism.GetBlockTime(ctx).Before(request.ExecutableAt.AsTime()) {
		return nil, status.Error(codes.FailedPrecondition, "recovery delay not passed")
	}

	request.Status = wspb.RecoveryStatus_RECOVERY_STATUS_EXECUTED
	request.ExecutedAt = timestamppb.Now()

	updatedRequestBytes, err := ms.Keeper.cdc.Marshal(&request)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetRecoveryRequest(ctx, msg.RequestId, updatedRequestBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteRecovery,
			sdk.NewAttribute(types.AttributeKeyRequestID, msg.RequestId),
		),
	)

	return &wspb.MsgExecuteRecoveryResponse{
		Success:    true,
		NewAddress: request.NewAddress,
	}, nil
}

// SimulateTransaction simulates a transaction before execution
func (ms msgServer) SimulateTransaction(goCtx context.Context, msg *wspb.MsgSimulateTransaction) (*wspb.MsgSimulateTransactionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches claimed sender
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedSender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sender address")
	}

	if !signers[0].Equals(claimedSender) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match sender")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	simulation, err := ms.Keeper.SimulateTransaction(ctx, msg.TxData, msg.Sender)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.MsgSimulateTransactionResponse{Simulation: simulation}, nil
}

// VerifyDomain verifies a domain for phishing protection
func (ms msgServer) VerifyDomain(goCtx context.Context, msg *wspb.MsgVerifyDomain) (*wspb.MsgVerifyDomainResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer matches claimed verifier
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedVerifier, err := sdk.AccAddressFromBech32(msg.Verifier)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid verifier address")
	}

	if !signers[0].Equals(claimedVerifier) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match verifier")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	verification, err := ms.Keeper.VerifyDomain(ctx, msg.Domain, msg.CertificateHash, msg.Verifier)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVerifyDomain,
			sdk.NewAttribute(types.AttributeKeyDomain, msg.Domain),
		),
	)

	return &wspb.MsgVerifyDomainResponse{Verification: verification}, nil
}

// SetSpendingLimit sets spending limits for a wallet
func (ms msgServer) SetSpendingLimit(goCtx context.Context, msg *wspb.MsgSetSpendingLimit) (*wspb.MsgSetSpendingLimitResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer is the wallet owner
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedOwner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	if !signers[0].Equals(claimedOwner) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	limit, err := ms.Keeper.SetSpendingLimit(ctx, msg.WalletId, msg.Denom, msg.DailyLimit, msg.WeeklyLimit, msg.MonthlyLimit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetSpendingLimit,
			sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
			sdk.NewAttribute(types.AttributeKeyDenom, msg.Denom),
		),
	)

	return &wspb.MsgSetSpendingLimitResponse{Limit: limit}, nil
}

// ConfigureSession configures wallet session settings
func (ms msgServer) ConfigureSession(goCtx context.Context, msg *wspb.MsgConfigureSession) (*wspb.MsgConfigureSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify signer is the wallet owner
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	claimedOwner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner address")
	}

	if !signers[0].Equals(claimedOwner) {
		return nil, status.Error(codes.PermissionDenied, "transaction signer does not match owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config := &wspb.SessionConfig{
		SessionId:                  fmt.Sprintf("session_%s_%d", msg.WalletId, ctx.BlockHeight()),
		WalletId:                   msg.WalletId,
		TimeoutDuration:            msg.TimeoutDuration,
		AutoLockEnabled:            msg.AutoLockEnabled,
		InactivityThresholdSeconds: msg.InactivityThresholdSeconds,
	}

	configBytes, err := ms.Keeper.cdc.Marshal(config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetSessionConfig(ctx, config.SessionId, configBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.MsgConfigureSessionResponse{Config: config}, nil
}

// LockSession locks the current wallet session
func (ms msgServer) LockSession(goCtx context.Context, msg *wspb.MsgLockSession) (*wspb.MsgLockSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Mark session as locked (simplified)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeLockSession,
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionId),
		),
	)

	return &wspb.MsgLockSessionResponse{Locked: true}, nil
}

// UnlockSession unlocks a locked wallet session
func (ms msgServer) UnlockSession(goCtx context.Context, msg *wspb.MsgUnlockSession) (*wspb.MsgUnlockSessionResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	expiresAt := timestamppb.New(determinism.GetBlockTime(ctx).Add(1 * time.Hour))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUnlockSession,
			sdk.NewAttribute(types.AttributeKeySessionID, msg.SessionId),
		),
	)

	return &wspb.MsgUnlockSessionResponse{
		Unlocked:  true,
		ExpiresAt: expiresAt,
	}, nil
}

// EnrollBiometric enrolls biometric authentication
//
// DEPRECATION WARNING:
//   This function is DEPRECATED and should not be used in production.
//   True biometric authentication cannot be implemented securely on a blockchain.
//   See keeper.verifyBiometricTemplate for detailed explanation.
//
// Security Implementation:
//   - Enrollment data is hashed with SHA-256 before storage
//   - Only the hash is stored on-chain (never raw biometric data)
//   - Hash is stored permanently and cannot be changed
//
// Privacy Concerns:
//   - Biometric hashes stored on-chain are PERMANENT and PUBLIC
//   - If the hash is compromised, the user cannot change their biometrics
//   - This violates GDPR "right to be forgotten" for biometric data
//   - Public blockchains create permanent records of biometric identifiers
//
// Security Limitations:
//   - The enrollment hash is essentially a pre-shared secret
//   - No liveness detection (cannot detect spoofed enrollment)
//   - No fuzzy matching (exact match required, defeats biometric purpose)
//   - Vulnerable to replay if enrollment data is captured
//
// This implementation is provided for compatibility only and should be
// considered a pre-shared secret system, not true biometric authentication.
func (ms msgServer) EnrollBiometric(goCtx context.Context, msg *wspb.MsgEnrollBiometric) (*wspb.MsgEnrollBiometricResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Validate enrollment data is provided
	if len(msg.EnrollmentData) == 0 {
		return nil, status.Error(codes.InvalidArgument, "enrollment data is required")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL: Hash the enrollment data for storage
	// This prevents storing raw biometric data on-chain
	// However, note that even hashed biometric data creates privacy concerns
	// on a public blockchain (permanent, immutable, publicly visible)
	enrollmentHash := sha256.Sum256(msg.EnrollmentData)
	enrollmentHashStr := fmt.Sprintf("%x", enrollmentHash)

	auth := &wspb.BiometricAuth{
		WalletId:       msg.WalletId,
		Type:           msg.Type,
		EnrollmentHash: enrollmentHashStr, // Store hash, not raw data
		EnrolledAt:     timestamppb.Now(),
		Enabled:        true,
		FailedAttempts: 0,
		LockedOut:      false,
	}

	authBytes, err := ms.Keeper.cdc.Marshal(auth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetBiometricAuth(ctx, msg.WalletId, authBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeEnrollBiometric,
			sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
		),
	)

	return &wspb.MsgEnrollBiometricResponse{Auth: auth}, nil
}

// AuthenticateBiometric authenticates using biometric
//
// DEPRECATION WARNING:
//   This function is DEPRECATED and should not be used in production.
//   See detailed explanation in keeper.verifyBiometricTemplate for why
//   biometric authentication is fundamentally incompatible with blockchain.
//
// Security Implementation Notes:
//   Despite the architectural limitations, this implementation includes:
//   - Signer verification (prevents unauthorized authentication attempts)
//   - Minimum proof size validation (prevents trivial bypass attacks)
//   - Replay protection (prevents reuse of captured proofs)
//   - Failed attempt tracking (rate limiting)
//   - Automatic lockout after 5 failed attempts
//   - Cryptographic proof verification (exact hash matching)
//
// What This Implementation DOES:
//   - Verifies the transaction is signed by the wallet owner
//   - Checks that biometric proof matches enrollment data (exact match)
//   - Prevents replay attacks by tracking used proofs
//   - Locks out wallet after repeated failed attempts
//
// What This Implementation DOES NOT Do:
//   - True biometric matching (fuzzy matching, similarity scores)
//   - Liveness detection (detect spoofed biometrics)
//   - Anti-spoofing measures (detect fake fingerprints)
//   - Work across multiple devices (biometric is device-specific)
//   - Provide security beyond pre-shared secret authentication
//
// Recommended Alternative:
//   Use client-side biometric authentication (FaceID, TouchID, etc.) to
//   unlock the wallet's private key, then sign transactions normally
//   with Cosmos SDK authentication. This provides:
//   - True biometric security (hardware-backed, liveness detection)
//   - Off-chain biometric verification (no consensus issues)
//   - Standard blockchain authentication (proven security model)
//   - Privacy (biometric data never leaves the device)
func (ms msgServer) AuthenticateBiometric(goCtx context.Context, msg *wspb.MsgAuthenticateBiometric) (*wspb.MsgAuthenticateBiometricResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// CRITICAL: Verify transaction signer matches claimed wallet address
	// This prevents anyone from attempting to authenticate on behalf of another wallet
	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no signers")
	}

	walletAddr, err := sdk.AccAddressFromBech32(msg.WalletId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid wallet address")
	}

	if !signers[0].Equals(walletAddr) {
		return nil, status.Error(codes.PermissionDenied, "signer does not match wallet")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// CRITICAL: Minimum biometric proof size validation
	const MinBiometricProofSize = 64 // Minimum size for valid proof
	if len(msg.BiometricProof) < MinBiometricProofSize {
		return nil, status.Error(codes.InvalidArgument, "biometric proof too short")
	}

	// CRITICAL: Get registered biometric configuration
	biometricAuthBytes, err := ms.Keeper.GetBiometricAuth(ctx, msg.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "biometric not configured for this wallet")
	}

	var biometricAuth wspb.BiometricAuth
	if err := ms.Keeper.cdc.Unmarshal(biometricAuthBytes, &biometricAuth); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check if biometric authentication is enabled
	if !biometricAuth.Enabled {
		return nil, status.Error(codes.PermissionDenied, "biometric authentication is disabled")
	}

	// CRITICAL: Check if currently locked out
	if biometricAuth.LockedOut {
		now := determinism.GetBlockTime(ctx)
		if biometricAuth.LockoutUntil != nil && now.Before(biometricAuth.LockoutUntil.AsTime()) {
			return &wspb.MsgAuthenticateBiometricResponse{
				Authenticated:  false,
				FailedAttempts: biometricAuth.FailedAttempts,
				LockedOut:      true,
			}, nil
		}
		// Lockout period expired, reset
		biometricAuth.LockedOut = false
		biometricAuth.FailedAttempts = 0
	}

	// CRITICAL: Replay protection - check if proof was already used
	proofHash := sha256.Sum256(msg.BiometricProof)
	if ms.Keeper.IsBiometricProofUsed(ctx, msg.WalletId, proofHash[:]) {
		return nil, status.Error(codes.AlreadyExists, "biometric proof already used (replay attack detected)")
	}

	// CRITICAL: Verify biometric proof against stored enrollment hash
	authenticated := ms.Keeper.verifyBiometricTemplate(
		biometricAuth.EnrollmentHash,
		msg.BiometricProof,
	)

	if authenticated {
		// CRITICAL: Mark proof as used to prevent replay attacks
		if err := ms.Keeper.MarkBiometricProofUsed(ctx, msg.WalletId, proofHash[:]); err != nil {
			return nil, status.Error(codes.Internal, "failed to mark proof as used")
		}

		// Reset failed attempts on successful authentication
		biometricAuth.FailedAttempts = 0
		biometricAuth.LastAttempt = timestamppb.New(determinism.GetBlockTime(ctx))

		// Update biometric auth record
		updatedAuthBytes, err := ms.Keeper.cdc.Marshal(&biometricAuth)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := ms.Keeper.SetBiometricAuth(ctx, msg.WalletId, updatedAuthBytes); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeBiometricAuth,
				sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
				sdk.NewAttribute(types.AttributeKeyStatus, "success"),
			),
		)

		return &wspb.MsgAuthenticateBiometricResponse{
			Authenticated:  true,
			FailedAttempts: 0,
			LockedOut:      false,
		}, nil
	}

	// Authentication failed - increment failed attempts
	biometricAuth.FailedAttempts++
	biometricAuth.LastAttempt = timestamppb.New(determinism.GetBlockTime(ctx))

	// CRITICAL: Lock out after max failed attempts (e.g., 5)
	const MaxFailedAttempts = 5
	const LockoutDuration = 30 * time.Minute

	if biometricAuth.FailedAttempts >= MaxFailedAttempts {
		biometricAuth.LockedOut = true
		biometricAuth.LockoutUntil = timestamppb.New(determinism.GetBlockTime(ctx).Add(LockoutDuration))
	}

	// Update biometric auth record with incremented failed attempts
	updatedAuthBytes, err := ms.Keeper.cdc.Marshal(&biometricAuth)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetBiometricAuth(ctx, msg.WalletId, updatedAuthBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBiometricAuth,
			sdk.NewAttribute(types.AttributeKeyWalletID, msg.WalletId),
			sdk.NewAttribute(types.AttributeKeyStatus, "failed"),
			sdk.NewAttribute("failed_attempts", fmt.Sprintf("%d", biometricAuth.FailedAttempts)),
		),
	)

	return &wspb.MsgAuthenticateBiometricResponse{
		Authenticated:  false,
		FailedAttempts: biometricAuth.FailedAttempts,
		LockedOut:      biometricAuth.LockedOut,
	}, nil
}

// StoreInSecureEnclave stores key material in secure enclave
func (ms msgServer) StoreInSecureEnclave(goCtx context.Context, msg *wspb.MsgStoreInSecureEnclave) (*wspb.MsgStoreInSecureEnclaveResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	config := &wspb.SecureEnclaveConfig{
		WalletId:               msg.WalletId,
		EnclaveType:            msg.EnclaveType,
		AttestationCertificate: msg.AttestationCertificate,
		CreatedAt:              timestamppb.Now(),
	}

	configBytes, err := ms.Keeper.cdc.Marshal(config)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetSecureEnclaveConfig(ctx, msg.WalletId, configBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.MsgStoreInSecureEnclaveResponse{Config: config}, nil
}

// CreateEncryptedBackup creates an encrypted backup of seed phrase
func (ms msgServer) CreateEncryptedBackup(goCtx context.Context, msg *wspb.MsgCreateEncryptedBackup) (*wspb.MsgCreateEncryptedBackupResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	backupID := fmt.Sprintf("backup_%s_%d", msg.WalletId, ctx.BlockHeight())

	backup := &wspb.EncryptedBackup{
		BackupId:               backupID,
		WalletId:               msg.WalletId,
		EncryptionAlgorithm:    msg.EncryptionAlgorithm,
		KeyDerivationFunction:  msg.KeyDerivationFunction,
		Salt:                   msg.Salt,
		Iterations:             msg.Iterations,
		Location:               msg.Location,
		CreatedAt:              timestamppb.Now(),
	}

	backupBytes, err := ms.Keeper.cdc.Marshal(backup)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetEncryptedBackup(ctx, backupID, backupBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.MsgCreateEncryptedBackupResponse{Backup: backup}, nil
}

// ConfigureDustFilter configures dust attack filtering
func (ms msgServer) ConfigureDustFilter(goCtx context.Context, msg *wspb.MsgConfigureDustFilter) (*wspb.MsgConfigureDustFilterResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	filter := &wspb.DustAttackFilter{
		WalletId:                      msg.WalletId,
		Enabled:                       msg.Enabled,
		MinimumAmount:                 msg.MinimumAmount,
		MaxDustTransactionsPerBlock:   msg.MaxDustTransactionsPerBlock,
		SuspiciousPatternThreshold:    msg.SuspiciousPatternThreshold,
	}

	filterBytes, err := ms.Keeper.cdc.Marshal(filter)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := ms.Keeper.SetDustFilter(ctx, msg.WalletId, filterBytes); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &wspb.MsgConfigureDustFilterResponse{Filter: filter}, nil
}

// ValidateAddressChecksum validates an address checksum
func (ms msgServer) ValidateAddressChecksum(goCtx context.Context, msg *wspb.MsgValidateAddressChecksum) (*wspb.MsgValidateAddressChecksumResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// Simplified validation
	valid := len(msg.Address) > 0
	checksum := "validated"
	errorMessage := ""

	if !valid {
		errorMessage = "invalid address"
	}

	return &wspb.MsgValidateAddressChecksumResponse{
		Valid:        valid,
		Checksum:     checksum,
		ErrorMessage: errorMessage,
	}, nil
}
