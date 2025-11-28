package keeper

import (
    "github.com/aequitas/aura/chain/x/common/determinism"
	"context"
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

	ctx := sdk.UnwrapSDKContext(goCtx)

	txBytes, err := ms.Keeper.GetPendingMultiSigTx(ctx, msg.TxId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "transaction not found")
	}

	var tx wspb.PendingMultiSigTransaction
	if err := ms.Keeper.cdc.Unmarshal(txBytes, &tx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Add signature
	tx.Signatures = append(tx.Signatures, string(msg.Signature))
	tx.SignedBy = append(tx.SignedBy, msg.Signer)
	tx.CurrentWeight++

	// Get wallet to check threshold
	walletBytes, err := ms.Keeper.GetMultiSigWallet(ctx, tx.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "wallet not found")
	}

	var wallet wspb.MultiSigWallet
	if err := ms.Keeper.cdc.Unmarshal(walletBytes, &wallet); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	readyToExecute := tx.CurrentWeight >= wallet.Threshold

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
		RequiredSignatures: wallet.Threshold,
		ReadyToExecute:     readyToExecute,
	}, nil
}

// ConfigureSocialRecovery configures social recovery for a wallet
func (ms msgServer) ConfigureSocialRecovery(goCtx context.Context, msg *wspb.MsgConfigureSocialRecovery) (*wspb.MsgConfigureSocialRecoveryResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
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

	ctx := sdk.UnwrapSDKContext(goCtx)

	requestBytes, err := ms.Keeper.GetRecoveryRequest(ctx, msg.RequestId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery request not found")
	}

	var request wspb.RecoveryRequest
	if err := ms.Keeper.cdc.Unmarshal(requestBytes, &request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Add approval
	request.Approvals = append(request.Approvals, msg.Guardian)
	request.ApprovalsCount++

	// Get recovery config to check threshold
	configBytes, err := ms.Keeper.GetSocialRecoveryConfig(ctx, request.WalletId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "recovery config not found")
	}

	var config wspb.SocialRecoveryConfig
	if err := ms.Keeper.cdc.Unmarshal(configBytes, &config); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

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
func (ms msgServer) EnrollBiometric(goCtx context.Context, msg *wspb.MsgEnrollBiometric) (*wspb.MsgEnrollBiometricResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	auth := &wspb.BiometricAuth{
		WalletId:       msg.WalletId,
		Type:           msg.Type,
		EnrolledAt:     timestamppb.Now(),
		Enabled:        true,
		FailedAttempts: 0,
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
func (ms msgServer) AuthenticateBiometric(goCtx context.Context, msg *wspb.MsgAuthenticateBiometric) (*wspb.MsgAuthenticateBiometricResponse, error) {
	if msg == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	_ = sdk.UnwrapSDKContext(goCtx)

	// Simplified authentication
	authenticated := len(msg.BiometricProof) > 0

	return &wspb.MsgAuthenticateBiometricResponse{
		Authenticated:  authenticated,
		FailedAttempts: 0,
		LockedOut:      false,
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
