package walletsecurity

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/protobuf/proto"

	"github.com/aequitas/aura/chain/x/walletsecurity/keeper"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

var (
	_ appmodule.AppModule = AppModule{}
	_ module.HasServices  = AppModule{}
)

// AppModule implements the AppModule interface for the wallet security module
type AppModule struct {
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// Name returns the wallet security module's name
func (AppModule) Name() string {
	return "walletsecurity"
}

// RegisterLegacyAminoCodec registers the wallet security module's types on the LegacyAmino codec
func (AppModule) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the wallet security module
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// TODO: Uncomment when gateway code is generated
	// wsproto.RegisterQueryHandlerClient(context.Background(), mux, wsproto.NewQueryClient(clientCtx))
	_ = clientCtx
	_ = mux
}

// RegisterInterfaces registers the wallet security module's interface types
func (AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {}

// ConsensusVersion implements AppModule/ConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	wsproto.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(am.keeper))
	wsproto.RegisterQueryServer(cfg.QueryServer(), NewQueryServerImpl(am.keeper))
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface
func (am AppModule) IsAppModule() {}

// MsgServerImpl implements the wallet security Msg service
type MsgServerImpl struct {
	wsproto.UnimplementedMsgServer
	keeper keeper.Keeper
}

// NewMsgServerImpl creates a new MsgServerImpl
func NewMsgServerImpl(keeper keeper.Keeper) *MsgServerImpl {
	return &MsgServerImpl{keeper: keeper}
}

// QueryServerImpl implements the wallet security Query service
type QueryServerImpl struct {
	wsproto.UnimplementedQueryServer
	keeper keeper.Keeper
}

// NewQueryServerImpl creates a new QueryServerImpl
func NewQueryServerImpl(keeper keeper.Keeper) *QueryServerImpl {
	return &QueryServerImpl{keeper: keeper}
}

// Implement Msg service methods
func (m *MsgServerImpl) RegisterHardwareWallet(
	ctx context.Context,
	req *wsproto.MsgRegisterHardwareWallet,
) (*wsproto.MsgRegisterHardwareWalletResponse, error) {
	config, err := m.keeper.RegisterHardwareWallet(
		ctx,
		req.Address,
		req.Type,
		req.DeviceId,
		req.FirmwareVersion,
		req.DerivationPath,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgRegisterHardwareWalletResponse{
		WalletId: config.WalletId,
		Config:   config,
	}, nil
}

func (m *MsgServerImpl) CreateMultiSigWallet(
	ctx context.Context,
	req *wsproto.MsgCreateMultiSigWallet,
) (*wsproto.MsgCreateMultiSigWalletResponse, error) {
	wallet, err := m.keeper.CreateMultiSigWallet(
		ctx,
		req.Creator,
		req.Signers,
		req.Threshold,
		req.SignerWeights,
		req.WeightThreshold,
		req.TimeLock,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgCreateMultiSigWalletResponse{
		WalletId: wallet.WalletId,
		Wallet:   wallet,
	}, nil
}

func (m *MsgServerImpl) SignMultiSigTransaction(
	ctx context.Context,
	req *wsproto.MsgSignMultiSigTransaction,
) (*wsproto.MsgSignMultiSigTransactionResponse, error) {
	ready, err := m.keeper.SignMultiSigTransaction(
		ctx,
		req.TxId,
		req.Signer,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgSignMultiSigTransactionResponse{
		ReadyToExecute: ready,
	}, nil
}

func (m *MsgServerImpl) ConfigureSocialRecovery(
	ctx context.Context,
	req *wsproto.MsgConfigureSocialRecovery,
) (*wsproto.MsgConfigureSocialRecoveryResponse, error) {
	config, err := m.keeper.ConfigureSocialRecovery(
		ctx,
		req.WalletId,
		req.Guardians,
		req.RecoveryThreshold,
		req.RecoveryDelay,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgConfigureSocialRecoveryResponse{
		Config: config,
	}, nil
}

func (m *MsgServerImpl) InitiateRecovery(
	ctx context.Context,
	req *wsproto.MsgInitiateRecovery,
) (*wsproto.MsgInitiateRecoveryResponse, error) {
	request, err := m.keeper.InitiateRecovery(
		ctx,
		req.WalletId,
		req.NewAddress,
		req.Initiator,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgInitiateRecoveryResponse{
		RequestId: request.RequestId,
		Request:   request,
	}, nil
}

func (m *MsgServerImpl) ApproveRecovery(
	ctx context.Context,
	req *wsproto.MsgApproveRecovery,
) (*wsproto.MsgApproveRecoveryResponse, error) {
	ready, err := m.keeper.ApproveRecovery(
		ctx,
		req.RequestId,
		req.Guardian,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgApproveRecoveryResponse{
		ReadyToExecute: ready,
	}, nil
}

func (m *MsgServerImpl) ExecuteRecovery(
	ctx context.Context,
	req *wsproto.MsgExecuteRecovery,
) (*wsproto.MsgExecuteRecoveryResponse, error) {
	err := m.keeper.ExecuteRecovery(ctx, req.RequestId)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgExecuteRecoveryResponse{
		Success: true,
	}, nil
}

func (m *MsgServerImpl) SimulateTransaction(
	ctx context.Context,
	req *wsproto.MsgSimulateTransaction,
) (*wsproto.MsgSimulateTransactionResponse, error) {
	simulation, err := m.keeper.SimulateTransaction(
		ctx,
		req.TxData,
		req.Sender,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgSimulateTransactionResponse{
		Simulation: simulation,
	}, nil
}

func (m *MsgServerImpl) VerifyDomain(
	ctx context.Context,
	req *wsproto.MsgVerifyDomain,
) (*wsproto.MsgVerifyDomainResponse, error) {
	verification, err := m.keeper.VerifyDomain(
		ctx,
		req.Domain,
		req.CertificateHash,
		req.Verifier,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgVerifyDomainResponse{
		Verification: verification,
	}, nil
}

func (m *MsgServerImpl) SetSpendingLimit(
	ctx context.Context,
	req *wsproto.MsgSetSpendingLimit,
) (*wsproto.MsgSetSpendingLimitResponse, error) {
	limit, err := m.keeper.SetSpendingLimit(
		ctx,
		req.WalletId,
		req.Denom,
		req.DailyLimit,
		req.WeeklyLimit,
		req.MonthlyLimit,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgSetSpendingLimitResponse{
		Limit: limit,
	}, nil
}

func (m *MsgServerImpl) ConfigureSession(
	ctx context.Context,
	req *wsproto.MsgConfigureSession,
) (*wsproto.MsgConfigureSessionResponse, error) {
	config, err := m.keeper.ConfigureSession(
		ctx,
		req.WalletId,
		req.TimeoutDuration,
		req.AutoLockEnabled,
		req.InactivityThresholdSeconds,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgConfigureSessionResponse{
		Config: config,
	}, nil
}

func (m *MsgServerImpl) LockSession(
	ctx context.Context,
	req *wsproto.MsgLockSession,
) (*wsproto.MsgLockSessionResponse, error) {
	err := m.keeper.LockSession(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgLockSessionResponse{
		Locked: true,
	}, nil
}

func (m *MsgServerImpl) UnlockSession(
	ctx context.Context,
	req *wsproto.MsgUnlockSession,
) (*wsproto.MsgUnlockSessionResponse, error) {
	err := m.keeper.UnlockSession(ctx, req.SessionId, req.AuthenticationProof)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgUnlockSessionResponse{
		Unlocked: true,
	}, nil
}

func (m *MsgServerImpl) EnrollBiometric(
	ctx context.Context,
	req *wsproto.MsgEnrollBiometric,
) (*wsproto.MsgEnrollBiometricResponse, error) {
	auth, err := m.keeper.EnrollBiometric(
		ctx,
		req.WalletId,
		req.Type,
		req.EnrollmentData,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgEnrollBiometricResponse{
		Auth: auth,
	}, nil
}

func (m *MsgServerImpl) AuthenticateBiometric(
	ctx context.Context,
	req *wsproto.MsgAuthenticateBiometric,
) (*wsproto.MsgAuthenticateBiometricResponse, error) {
	authenticated, err := m.keeper.AuthenticateBiometric(
		ctx,
		req.WalletId,
		req.BiometricProof,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgAuthenticateBiometricResponse{
		Authenticated: authenticated,
	}, nil
}

func (m *MsgServerImpl) StoreInSecureEnclave(
	ctx context.Context,
	req *wsproto.MsgStoreInSecureEnclave,
) (*wsproto.MsgStoreInSecureEnclaveResponse, error) {
	config, err := m.keeper.StoreInSecureEnclave(
		ctx,
		req.WalletId,
		req.EnclaveType,
		req.EncryptedKeyMaterial,
		req.AttestationCertificate,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgStoreInSecureEnclaveResponse{
		Config: config,
	}, nil
}

func (m *MsgServerImpl) CreateEncryptedBackup(
	ctx context.Context,
	req *wsproto.MsgCreateEncryptedBackup,
) (*wsproto.MsgCreateEncryptedBackupResponse, error) {
	backup, err := m.keeper.CreateEncryptedBackup(
		ctx,
		req.WalletId,
		req.EncryptedSeed,
		req.EncryptionAlgorithm,
		req.KeyDerivationFunction,
		req.Salt,
		req.Iterations,
		req.Location,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgCreateEncryptedBackupResponse{
		Backup: backup,
	}, nil
}

func (m *MsgServerImpl) ConfigureDustFilter(
	ctx context.Context,
	req *wsproto.MsgConfigureDustFilter,
) (*wsproto.MsgConfigureDustFilterResponse, error) {
	filter, err := m.keeper.ConfigureDustFilter(
		ctx,
		req.WalletId,
		req.Enabled,
		req.MinimumAmount,
		req.MaxDustTransactionsPerBlock,
		req.SuspiciousPatternThreshold,
	)
	if err != nil {
		return nil, err
	}

	return &wsproto.MsgConfigureDustFilterResponse{
		Filter: filter,
	}, nil
}

func (m *MsgServerImpl) ValidateAddressChecksum(
	ctx context.Context,
	req *wsproto.MsgValidateAddressChecksum,
) (*wsproto.MsgValidateAddressChecksumResponse, error) {
	valid, checksum, err := m.keeper.ValidateAddressChecksum(
		ctx,
		req.Address,
		req.Algorithm,
	)
	if err != nil {
		return &wsproto.MsgValidateAddressChecksumResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &wsproto.MsgValidateAddressChecksumResponse{
		Valid:    valid,
		Checksum: checksum,
	}, nil
}

// Implement Query service methods
func (q *QueryServerImpl) GetHardwareWallet(
	ctx context.Context,
	req *wsproto.QueryGetHardwareWalletRequest,
) (*wsproto.QueryGetHardwareWalletResponse, error) {
	configBytes, err := q.keeper.GetHardwareWallet(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	var config wsproto.HardwareWalletConfig
	if err := proto.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetHardwareWalletResponse{
		Config: &config,
	}, nil
}

func (q *QueryServerImpl) GetMultiSigWallet(
	ctx context.Context,
	req *wsproto.QueryGetMultiSigWalletRequest,
) (*wsproto.QueryGetMultiSigWalletResponse, error) {
	walletBytes, err := q.keeper.GetMultiSigWallet(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	var wallet wsproto.MultiSigWallet
	if err := proto.Unmarshal(walletBytes, &wallet); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetMultiSigWalletResponse{
		Wallet: &wallet,
	}, nil
}

func (q *QueryServerImpl) GetPendingMultiSigTx(
	ctx context.Context,
	req *wsproto.QueryGetPendingMultiSigTxRequest,
) (*wsproto.QueryGetPendingMultiSigTxResponse, error) {
	txBytes, err := q.keeper.GetPendingMultiSigTx(ctx, req.TxId)
	if err != nil {
		return nil, err
	}

	var tx wsproto.PendingMultiSigTransaction
	if err := proto.Unmarshal(txBytes, &tx); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetPendingMultiSigTxResponse{
		Tx: &tx,
	}, nil
}

func (q *QueryServerImpl) GetSocialRecoveryConfig(
	ctx context.Context,
	req *wsproto.QueryGetSocialRecoveryConfigRequest,
) (*wsproto.QueryGetSocialRecoveryConfigResponse, error) {
	configBytes, err := q.keeper.GetSocialRecoveryConfig(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	var config wsproto.SocialRecoveryConfig
	if err := proto.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetSocialRecoveryConfigResponse{
		Config: &config,
	}, nil
}

func (q *QueryServerImpl) GetRecoveryRequest(
	ctx context.Context,
	req *wsproto.QueryGetRecoveryRequestRequest,
) (*wsproto.QueryGetRecoveryRequestResponse, error) {
	requestBytes, err := q.keeper.GetRecoveryRequest(ctx, req.RequestId)
	if err != nil {
		return nil, err
	}

	var request wsproto.RecoveryRequest
	if err := proto.Unmarshal(requestBytes, &request); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetRecoveryRequestResponse{
		Request: &request,
	}, nil
}

func (q *QueryServerImpl) GetSpendingLimit(
	ctx context.Context,
	req *wsproto.QueryGetSpendingLimitRequest,
) (*wsproto.QueryGetSpendingLimitResponse, error) {
	limitBytes, err := q.keeper.GetSpendingLimit(ctx, req.WalletId, req.Denom)
	if err != nil {
		return nil, err
	}

	var limit wsproto.SpendingLimit
	if err := proto.Unmarshal(limitBytes, &limit); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetSpendingLimitResponse{
		Limit: &limit,
	}, nil
}

func (q *QueryServerImpl) GetSessionConfig(
	ctx context.Context,
	req *wsproto.QueryGetSessionConfigRequest,
) (*wsproto.QueryGetSessionConfigResponse, error) {
	configBytes, err := q.keeper.GetSessionConfig(ctx, req.SessionId)
	if err != nil {
		return nil, err
	}

	var config wsproto.SessionConfig
	if err := proto.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetSessionConfigResponse{
		Config: &config,
	}, nil
}

func (q *QueryServerImpl) GetSecurityMetrics(
	ctx context.Context,
	req *wsproto.QueryGetSecurityMetricsRequest,
) (*wsproto.QueryGetSecurityMetricsResponse, error) {
	metricsBytes, err := q.keeper.GetSecurityMetrics(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	var metrics wsproto.WalletSecurityMetrics
	if err := proto.Unmarshal(metricsBytes, &metrics); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetSecurityMetricsResponse{
		Metrics: &metrics,
	}, nil
}

func (q *QueryServerImpl) GetDomainVerification(
	ctx context.Context,
	req *wsproto.QueryGetDomainVerificationRequest,
) (*wsproto.QueryGetDomainVerificationResponse, error) {
	verificationBytes, err := q.keeper.GetDomainVerification(ctx, req.Domain)
	if err != nil {
		return nil, err
	}

	var verification wsproto.DomainVerification
	if err := proto.Unmarshal(verificationBytes, &verification); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetDomainVerificationResponse{
		Verification: &verification,
	}, nil
}

func (q *QueryServerImpl) GetDustFilter(
	ctx context.Context,
	req *wsproto.QueryGetDustFilterRequest,
) (*wsproto.QueryGetDustFilterResponse, error) {
	filterBytes, err := q.keeper.GetDustFilter(ctx, req.WalletId)
	if err != nil {
		return nil, err
	}

	var filter wsproto.DustAttackFilter
	if err := proto.Unmarshal(filterBytes, &filter); err != nil {
		return nil, err
	}

	return &wsproto.QueryGetDustFilterResponse{
		Filter: &filter,
	}, nil
}
