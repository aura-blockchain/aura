package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	walletsecproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ walletsecproto.MsgServer = msgServer{}

type msgServer struct {
	twalletsecproto.UnimplementedMsgServer
	Keeper *Keeper
}

func NewMsgServerImpl(keeper *Keeper) walletsecproto.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (ms msgServer) RegisterHardwareWallet(goCtx context.Context, msg *walletsecproto.MsgRegisterHardwareWallet) (*walletsecproto.MsgRegisterHardwareWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Address == "" || msg.DeviceId == "" {
		return nil, types.ErrInvalidInput.Wrap("address and device_id are required")
	}

	walletID, config, err := ms.Keeper.RegisterHardwareWallet(ctx, msg.Address, msg.Type, msg.DeviceId, msg.FirmwareVersion, msg.DerivationPath, msg.Signature)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgRegisterHardwareWalletResponse{
		WalletId: walletID,
		Config:   config,
	}, nil
}

func (ms msgServer) CreateMultiSigWallet(goCtx context.Context, msg *walletsecproto.MsgCreateMultiSigWallet) (*walletsecproto.MsgCreateMultiSigWalletResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(msg.Signers) < 2 || msg.Threshold < 1 || msg.Threshold > int32(len(msg.Signers)) {
		return nil, types.ErrInvalidInput.Wrap("invalid signers or threshold")
	}

	walletID, wallet, err := ms.Keeper.CreateMultiSigWallet(ctx, msg.Creator, msg.Signers, msg.Threshold, msg.SignerWeights, msg.WeightThreshold, msg.TimeLock)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgCreateMultiSigWalletResponse{
		WalletId: walletID,
		Wallet:   wallet,
	}, nil
}

func (ms msgServer) SignMultiSigTransaction(goCtx context.Context, msg *walletsecproto.MsgSignMultiSigTransaction) (*walletsecproto.MsgSignMultiSigTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	currentSigs, requiredSigs, ready, err := ms.Keeper.SignMultiSigTransaction(ctx, msg.TxId, msg.Signer, msg.Signature)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgSignMultiSigTransactionResponse{
		CurrentSignatures:  currentSigs,
		RequiredSignatures: requiredSigs,
		ReadyToExecute:     ready,
	}, nil
}

func (ms msgServer) ConfigureSocialRecovery(goCtx context.Context, msg *walletsecproto.MsgConfigureSocialRecovery) (*walletsecproto.MsgConfigureSocialRecoveryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(msg.Guardians) < 3 {
		return nil, types.ErrInvalidInput.Wrap("at least 3 guardians required")
	}

	config, err := ms.Keeper.ConfigureSocialRecovery(ctx, msg.WalletId, msg.Guardians, msg.RecoveryThreshold, msg.RecoveryDelay)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgConfigureSocialRecoveryResponse{Config: config}, nil
}

func (ms msgServer) InitiateRecovery(goCtx context.Context, msg *walletsecproto.MsgInitiateRecovery) (*walletsecproto.MsgInitiateRecoveryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	requestID, request, err := ms.Keeper.InitiateRecovery(ctx, msg.WalletId, msg.NewAddress, msg.Initiator)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgInitiateRecoveryResponse{
		RequestId: requestID,
		Request:   request,
	}, nil
}

func (ms msgServer) ApproveRecovery(goCtx context.Context, msg *walletsecproto.MsgApproveRecovery) (*walletsecproto.MsgApproveRecoveryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	count, required, ready, executableAt, err := ms.Keeper.ApproveRecovery(ctx, msg.RequestId, msg.Guardian, msg.Signature)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgApproveRecoveryResponse{
		ApprovalsCount:    count,
		RequiredApprovals: required,
		ReadyToExecute:    ready,
		ExecutableAt:      executableAt,
	}, nil
}

func (ms msgServer) ExecuteRecovery(goCtx context.Context, msg *walletsecproto.MsgExecuteRecovery) (*walletsecproto.MsgExecuteRecoveryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	success, newAddress, err := ms.Keeper.ExecuteRecovery(ctx, msg.RequestId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgExecuteRecoveryResponse{
		Success:    success,
		NewAddress: newAddress,
	}, nil
}

func (ms msgServer) SimulateTransaction(goCtx context.Context, msg *walletsecproto.MsgSimulateTransaction) (*walletsecproto.MsgSimulateTransactionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	simulation, err := ms.Keeper.SimulateTransaction(ctx, msg.TxData, msg.Sender)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgSimulateTransactionResponse{Simulation: simulation}, nil
}

func (ms msgServer) VerifyDomain(goCtx context.Context, msg *walletsecproto.MsgVerifyDomain) (*walletsecproto.MsgVerifyDomainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	verification, err := ms.Keeper.VerifyDomain(ctx, msg.Domain, msg.CertificateHash, msg.Verifier)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgVerifyDomainResponse{Verification: verification}, nil
}

func (ms msgServer) SetSpendingLimit(goCtx context.Context, msg *walletsecproto.MsgSetSpendingLimit) (*walletsecproto.MsgSetSpendingLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	limit, err := ms.Keeper.SetSpendingLimit(ctx, msg.WalletId, msg.Denom, msg.DailyLimit, msg.WeeklyLimit, msg.MonthlyLimit)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgSetSpendingLimitResponse{Limit: limit}, nil
}

func (ms msgServer) ConfigureSession(goCtx context.Context, msg *walletsecproto.MsgConfigureSession) (*walletsecproto.MsgConfigureSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := ms.Keeper.ConfigureSession(ctx, msg.WalletId, msg.TimeoutDuration, msg.AutoLockEnabled, msg.InactivityThresholdSeconds)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgConfigureSessionResponse{Config: config}, nil
}

func (ms msgServer) LockSession(goCtx context.Context, msg *walletsecproto.MsgLockSession) (*walletsecproto.MsgLockSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	locked, err := ms.Keeper.LockSession(ctx, msg.SessionId)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgLockSessionResponse{Locked: locked}, nil
}

func (ms msgServer) UnlockSession(goCtx context.Context, msg *walletsecproto.MsgUnlockSession) (*walletsecproto.MsgUnlockSessionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	unlocked, expiresAt, err := ms.Keeper.UnlockSession(ctx, msg.SessionId, msg.AuthenticationProof)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgUnlockSessionResponse{
		Unlocked:  unlocked,
		ExpiresAt: expiresAt,
	}, nil
}

func (ms msgServer) EnrollBiometric(goCtx context.Context, msg *walletsecproto.MsgEnrollBiometric) (*walletsecproto.MsgEnrollBiometricResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	auth, err := ms.Keeper.EnrollBiometric(ctx, msg.WalletId, msg.Type, msg.EnrollmentData)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgEnrollBiometricResponse{Auth: auth}, nil
}

func (ms msgServer) AuthenticateBiometric(goCtx context.Context, msg *walletsecproto.MsgAuthenticateBiometric) (*walletsecproto.MsgAuthenticateBiometricResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authenticated, failedAttempts, lockedOut, err := ms.Keeper.AuthenticateBiometric(ctx, msg.WalletId, msg.BiometricProof)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgAuthenticateBiometricResponse{
		Authenticated:  authenticated,
		FailedAttempts: failedAttempts,
		LockedOut:      lockedOut,
	}, nil
}

func (ms msgServer) StoreInSecureEnclave(goCtx context.Context, msg *walletsecproto.MsgStoreInSecureEnclave) (*walletsecproto.MsgStoreInSecureEnclaveResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	config, err := ms.Keeper.StoreInSecureEnclave(ctx, msg.WalletId, msg.EnclaveType, msg.EncryptedKeyMaterial, msg.AttestationCertificate)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgStoreInSecureEnclaveResponse{Config: config}, nil
}

func (ms msgServer) CreateEncryptedBackup(goCtx context.Context, msg *walletsecproto.MsgCreateEncryptedBackup) (*walletsecproto.MsgCreateEncryptedBackupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	backup, err := ms.Keeper.CreateEncryptedBackup(ctx, msg.WalletId, msg.EncryptedSeed, msg.EncryptionAlgorithm, msg.KeyDerivationFunction, msg.Salt, msg.Iterations, msg.Location)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgCreateEncryptedBackupResponse{Backup: backup}, nil
}

func (ms msgServer) ConfigureDustFilter(goCtx context.Context, msg *walletsecproto.MsgConfigureDustFilter) (*walletsecproto.MsgConfigureDustFilterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	filter, err := ms.Keeper.ConfigureDustFilter(ctx, msg.WalletId, msg.Enabled, msg.MinimumAmount, msg.MaxDustTransactionsPerBlock, msg.SuspiciousPatternThreshold)
	if err != nil {
		return nil, err
	}

	return &walletsecproto.MsgConfigureDustFilterResponse{Filter: filter}, nil
}

func (ms msgServer) ValidateAddressChecksum(goCtx context.Context, msg *walletsecproto.MsgValidateAddressChecksum) (*walletsecproto.MsgValidateAddressChecksumResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valid, checksum, errorMsg := ms.Keeper.ValidateAddressChecksum(ctx, msg.Address, msg.Algorithm)

	return &walletsecproto.MsgValidateAddressChecksumResponse{
		Valid:        valid,
		Checksum:     checksum,
		ErrorMessage: errorMsg,
	}, nil
}
