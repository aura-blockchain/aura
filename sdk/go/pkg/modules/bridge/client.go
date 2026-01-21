package bridge

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	bridgepb "github.com/aequitas/aura/proto/aura/bridge/v1beta1"
	"github.com/aura-chain/aura/sdk/go/client"
	"github.com/aura-chain/aura/sdk/go/pkg/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

// Client provides methods for interacting with the bridge module
type Client struct {
	auraClient  *client.Client
	grpcConn    *grpc.ClientConn
	queryClient bridgepb.QueryClient
	msgClient   bridgepb.MsgClient
}

// NewClient creates a new bridge client
func NewClient(auraClient *client.Client) *Client {
	grpcConn := auraClient.GetClientContext().GRPCClient
	return &Client{
		auraClient:  auraClient,
		grpcConn:    grpcConn,
		queryClient: bridgepb.NewQueryClient(grpcConn),
		msgClient:   bridgepb.NewMsgClient(grpcConn),
	}
}

// LockTokensParams contains parameters for locking tokens
type LockTokensParams struct {
	Sender      string
	TargetChain string
	Recipient   string
	Amount      sdk.Coin
}

// LockTokensResponse contains the response from locking tokens
type LockTokensResponse struct {
	TransferID          string
	EstimatedCompletion uint64
}

// LockTokens locks tokens on AURA for transfer to PAW or XAI
func (c *Client) LockTokens(ctx context.Context, params *LockTokensParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return nil, fmt.Errorf("sender is required")
	}
	if params.TargetChain == "" {
		return nil, fmt.Errorf("target chain is required")
	}
	if params.Recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	// Validate amount is positive
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("amount must be positive, got %s", params.Amount.String())
	}
	// Validate denom is not empty
	if params.Amount.Denom == "" {
		return nil, fmt.Errorf("coin denomination cannot be empty")
	}

	msg := &bridgepb.MsgLockTokens{
		Sender:      params.Sender,
		TargetChain: params.TargetChain,
		Recipient:   params.Recipient,
		Amount:      params.Amount,
	}

	addr, err := sdk.AccAddressFromBech32(params.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to lock tokens: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// MintTokensParams contains parameters for minting tokens
type MintTokensParams struct {
	Validator          string
	SourceChain        string
	SourceTxHash       string
	Recipient          string
	Amount             math.Int
	Denom              string
	ValidatorSignature []byte
}

// MintTokensResponse contains the response from minting tokens
type MintTokensResponse struct {
	Success      bool
	WrappedDenom string
}

// MintTokens mints wrapped tokens on AURA from PAW or XAI (validator-only)
func (c *Client) MintTokens(ctx context.Context, params *MintTokensParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Validator == "" {
		return nil, fmt.Errorf("validator is required")
	}
	if params.SourceChain == "" {
		return nil, fmt.Errorf("source chain is required")
	}
	if params.SourceTxHash == "" {
		return nil, fmt.Errorf("source transaction hash is required")
	}
	if params.Recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	if params.Denom == "" {
		return nil, fmt.Errorf("denomination is required")
	}
	// Validate amount is positive
	if params.Amount.IsNil() || !params.Amount.IsPositive() {
		return nil, fmt.Errorf("amount must be positive")
	}
	if len(params.ValidatorSignature) == 0 {
		return nil, fmt.Errorf("validator signature is required")
	}

	msg := &bridgepb.MsgMintTokens{
		Validator:          params.Validator,
		SourceChain:        params.SourceChain,
		SourceTxHash:       params.SourceTxHash,
		Recipient:          params.Recipient,
		Amount:             params.Amount,
		Denom:              params.Denom,
		ValidatorSignature: params.ValidatorSignature,
	}

	addr, err := sdk.AccAddressFromBech32(params.Validator)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to mint tokens: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// UnlockTokensParams contains parameters for unlocking tokens
type UnlockTokensParams struct {
	Sender              string
	SourceChain         string
	BurnTxHash          string
	Amount              math.Int
	Denom               string
	ValidatorSignatures [][]byte
}

// UnlockTokens unlocks tokens on AURA after burn proof from target chain
func (c *Client) UnlockTokens(ctx context.Context, params *UnlockTokensParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return nil, fmt.Errorf("sender is required")
	}

	msg := &bridgepb.MsgUnlockTokens{
		Sender:              params.Sender,
		SourceChain:         params.SourceChain,
		BurnTxHash:          params.BurnTxHash,
		Amount:              params.Amount,
		Denom:               params.Denom,
		ValidatorSignatures: params.ValidatorSignatures,
	}

	addr, err := sdk.AccAddressFromBech32(params.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock tokens: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// BurnTokensParams contains parameters for burning tokens
type BurnTokensParams struct {
	Sender      string
	TargetChain string
	Recipient   string
	Amount      sdk.Coin
}

// BurnTokens burns wrapped tokens on AURA to unlock on source chain
func (c *Client) BurnTokens(ctx context.Context, params *BurnTokensParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return nil, fmt.Errorf("sender is required")
	}
	if params.TargetChain == "" {
		return nil, fmt.Errorf("target chain is required")
	}
	if params.Recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	// Validate amount is positive
	if !params.Amount.IsPositive() {
		return nil, fmt.Errorf("amount must be positive, got %s", params.Amount.String())
	}
	// Validate denom is not empty
	if params.Amount.Denom == "" {
		return nil, fmt.Errorf("coin denomination cannot be empty")
	}

	msg := &bridgepb.MsgBurnTokens{
		Sender:      params.Sender,
		TargetChain: params.TargetChain,
		Recipient:   params.Recipient,
		Amount:      params.Amount,
	}

	addr, err := sdk.AccAddressFromBech32(params.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to burn tokens: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// LinkAddressParams contains parameters for linking addresses
type LinkAddressParams struct {
	AuraAddress  string
	PawAddress   string
	XaiAddress   string
	PawSignature []byte
	XaiSignature []byte
	Signer       string
}

// LinkAddress links addresses across chains for shared identity
func (c *Client) LinkAddress(ctx context.Context, params *LinkAddressParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.AuraAddress == "" {
		return nil, fmt.Errorf("AURA address is required")
	}

	msg := &bridgepb.MsgLinkAddress{
		AuraAddress:  params.AuraAddress,
		PawAddress:   params.PawAddress,
		XaiAddress:   params.XaiAddress,
		PawSignature: params.PawSignature,
		XaiSignature: params.XaiSignature,
		Signer:       params.Signer,
	}

	addr, err := sdk.AccAddressFromBech32(params.Signer)
	if err != nil {
		return nil, fmt.Errorf("invalid signer address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to link address: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// CrossChainSwapParams contains parameters for cross-chain swap
type CrossChainSwapParams struct {
	Sender          string
	SourceChain     string
	InputCoin       sdk.Coin
	TargetChain     string
	TargetDenom     string
	MinTargetAmount math.Int
	Recipient       string
	MaxSlippageBPS  uint64
}

// CrossChainSwap initiates cross-chain swap
func (c *Client) CrossChainSwap(ctx context.Context, params *CrossChainSwapParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Sender == "" {
		return nil, fmt.Errorf("sender is required")
	}
	if params.SourceChain == "" {
		return nil, fmt.Errorf("source chain is required")
	}
	if params.TargetChain == "" {
		return nil, fmt.Errorf("target chain is required")
	}
	if params.TargetDenom == "" {
		return nil, fmt.Errorf("target denomination is required")
	}
	if params.Recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	// Validate input coin amount is positive
	if !params.InputCoin.IsPositive() {
		return nil, fmt.Errorf("input amount must be positive, got %s", params.InputCoin.String())
	}
	// Validate input denom is not empty
	if params.InputCoin.Denom == "" {
		return nil, fmt.Errorf("input coin denomination cannot be empty")
	}
	// Validate min target amount is positive
	if params.MinTargetAmount.IsNil() || !params.MinTargetAmount.IsPositive() {
		return nil, fmt.Errorf("minimum target amount must be positive")
	}
	// Validate max slippage is within bounds (0-10000 basis points = 0-100%)
	if params.MaxSlippageBPS > 10000 {
		return nil, fmt.Errorf("max slippage must be between 0 and 10000 basis points, got %d", params.MaxSlippageBPS)
	}

	msg := &bridgepb.MsgCrossChainSwap{
		Sender:          params.Sender,
		SourceChain:     params.SourceChain,
		InputCoin:       params.InputCoin,
		TargetChain:     params.TargetChain,
		TargetDenom:     params.TargetDenom,
		MinTargetAmount: params.MinTargetAmount,
		Recipient:       params.Recipient,
		MaxSlippageBps:  params.MaxSlippageBPS,
	}

	addr, err := sdk.AccAddressFromBech32(params.Sender)
	if err != nil {
		return nil, fmt.Errorf("invalid sender address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to perform cross-chain swap: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// RelayTransferParams contains parameters for relaying transfer
type RelayTransferParams struct {
	Relayer      string
	TransferID   string
	TargetTxHash string
	Status       string
}

// RelayTransfer relays cross-chain transfer (relayer-only)
func (c *Client) RelayTransfer(ctx context.Context, params *RelayTransferParams) (*types.TxResponse, error) {
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Relayer == "" {
		return nil, fmt.Errorf("relayer is required")
	}

	msg := &bridgepb.MsgRelayTransfer{
		Relayer:      params.Relayer,
		TransferId:   params.TransferID,
		TargetTxHash: params.TargetTxHash,
		Status:       params.Status,
	}

	addr, err := sdk.AccAddressFromBech32(params.Relayer)
	if err != nil {
		return nil, fmt.Errorf("invalid relayer address: %w", err)
	}

	txResp, err := c.auraClient.SignAndBroadcast(ctx, addr.String(), msg)
	if err != nil {
		return nil, fmt.Errorf("failed to relay transfer: %w", err)
	}

	return &types.TxResponse{
		TxHash: txResp.TxHash,
		Height: txResp.Height,
		Code:   txResp.Code,
		RawLog: txResp.RawLog,
	}, nil
}

// GetBridgeTransfer gets a specific bridge transfer by ID
func (c *Client) GetBridgeTransfer(ctx context.Context, id string) (*bridgepb.CrossChainTransfer, error) {
	if id == "" {
		return nil, fmt.Errorf("transfer ID is required")
	}

	req := &bridgepb.QueryTransferRequest{
		TransferId: id,
	}

	resp, err := c.queryClient.Transfer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge transfer: %w", err)
	}

	return &resp.Transfer, nil
}

// GetBridgeTransfers gets all bridge transfers for an address
func (c *Client) GetBridgeTransfers(ctx context.Context, address string) ([]*bridgepb.CrossChainTransfer, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &bridgepb.QueryUserTransfersRequest{
		Address: address,
	}

	resp, err := c.queryClient.UserTransfers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge transfers: %w", err)
	}

	// Convert value slice to pointer slice
	transfers := make([]*bridgepb.CrossChainTransfer, len(resp.Transfers))
	for i := range resp.Transfers {
		transfers[i] = &resp.Transfers[i]
	}
	return transfers, nil
}

// GetBridgeParams gets the bridge module parameters
func (c *Client) GetBridgeParams(ctx context.Context) (*bridgepb.ChainConfig, error) {
	// Note: Bridge module doesn't have a Params query, using ChainConfig instead
	req := &bridgepb.QueryChainConfigRequest{
		ChainId: "",
	}

	resp, err := c.queryClient.ChainConfig(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge params: %w", err)
	}

	return &resp.Config, nil
}

// GetBridgeStats gets bridge statistics
func (c *Client) GetBridgeStats(ctx context.Context) (*bridgepb.QueryBridgeStatsResponse, error) {
	req := &bridgepb.QueryBridgeStatsRequest{}

	resp, err := c.queryClient.BridgeStats(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge stats: %w", err)
	}

	return resp, nil
}

// GetLinkedAddresses gets linked addresses for a given address
func (c *Client) GetLinkedAddresses(ctx context.Context, address string) (*bridgepb.SharedIdentity, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	req := &bridgepb.QuerySharedIdentityRequest{
		Address: address,
	}

	resp, err := c.queryClient.SharedIdentity(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked addresses: %w", err)
	}

	return &resp.Identity, nil
}
