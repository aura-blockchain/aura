package faucet

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	sdkmath "cosmossdk.io/math"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aura-chain/aura/faucet/pkg/config"
	"github.com/aura-chain/aura/faucet/pkg/database"
)

const (
	faucetAccountName = "faucet"
	defaultTimeout    = 30 * time.Second
)

// EncodingConfig mirrors the minimal Cosmos SDK encoding config needed for signing.
type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
}

// Service handles faucet operations
type Service struct {
	cfg            *config.Config
	db             *database.DB
	httpClient     *http.Client
	grpcConn       *grpc.ClientConn
	authClient     authtypes.QueryClient
	bankClient     banktypes.QueryClient
	encodingConfig EncodingConfig
	keyring        keyring.Keyring
	fromAddress    sdk.AccAddress
	fromName       string
	feeAmount      sdk.Coins
	txFactory      tx.Factory
	signMutex      sync.Mutex
}

// SendRequest represents a token send request
type SendRequest struct {
	Recipient string
	Amount    int64
	IPAddress string
}

// SendResponse represents a token send response
type SendResponse struct {
	TxHash    string
	Recipient string
	Amount    int64
}

// NodeStatus represents blockchain node status
type NodeStatus struct {
	NodeInfo struct {
		Network string `json:"network"`
		Version string `json:"version"`
	} `json:"node_info"`
	SyncInfo struct {
		LatestBlockHeight string `json:"latest_block_height"`
		CatchingUp        bool   `json:"catching_up"`
	} `json:"sync_info"`
}

// Balance represents account balance
type Balance struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
}

// NewService creates a new faucet service
func NewService(cfg *config.Config, db *database.DB) (*Service, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	initSDKConfig()

	encodingConfig := newEncodingConfig()

	fromName := faucetAccountName
	kr := keyring.NewInMemory(encodingConfig.Codec)
	record, err := kr.NewAccount(fromName, cfg.FaucetMnemonic, "", hd.CreateHDPath(sdk.CoinType, 0, 0).String(), hd.Secp256k1)
	if err != nil {
		return nil, fmt.Errorf("failed to derive faucet account: %w", err)
	}

	fromAddr, err := record.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get faucet address from keyring: %w", err)
	}

	// Use derived address if not explicitly set
	if cfg.FaucetAddress == "" {
		cfg.FaucetAddress = fromAddr.String()
	}

	if cfg.FaucetAddress != fromAddr.String() {
		return nil, fmt.Errorf("configured faucet address does not match derived address")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcConn, err := grpc.DialContext(ctx, cfg.NodeGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC endpoint %s: %w", cfg.NodeGRPC, err)
	}

	authClient := authtypes.NewQueryClient(grpcConn)
	bankClient := banktypes.NewQueryClient(grpcConn)

	gasPrice, err := sdk.ParseDecCoin(cfg.GasPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gas price %s: %w", cfg.GasPrice, err)
	}

	feeAmount := gasPrice.Amount.MulInt64(int64(cfg.GasLimit)).Ceil().RoundInt()
	feeCoins := sdk.NewCoins(sdk.NewCoin(gasPrice.Denom, feeAmount))

	txFactory := tx.Factory{}.
		WithChainID(cfg.ChainID).
		WithTxConfig(encodingConfig.TxConfig).
		WithKeybase(kr).
		WithGas(cfg.GasLimit).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	httpClient := &http.Client{
		Timeout: defaultTimeout,
	}

	return &Service{
		cfg:            cfg,
		db:             db,
		httpClient:     httpClient,
		grpcConn:       grpcConn,
		authClient:     authClient,
		bankClient:     bankClient,
		encodingConfig: encodingConfig,
		keyring:        kr,
		fromAddress:    fromAddr,
		fromName:       fromName,
		feeAmount:      feeCoins,
		txFactory:      txFactory,
	}, nil
}

// SendTokens sends tokens to a recipient
func (s *Service) SendTokens(req *SendRequest) (*SendResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	log.WithFields(log.Fields{
		"recipient": req.Recipient,
		"amount":    req.Amount,
		"ip":        req.IPAddress,
	}).Info("Sending tokens")

	dbReq, err := s.db.CreateRequest(req.Recipient, req.IPAddress, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to create request record: %w", err)
	}

	recipientAddr, err := sdk.AccAddressFromBech32(req.Recipient)
	if err != nil {
		_ = s.db.UpdateRequestFailed(dbReq.ID, "invalid recipient address")
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}

	msg := banktypes.NewMsgSend(s.fromAddress, recipientAddr, sdk.NewCoins(sdk.NewCoin(s.cfg.Denom, sdkmath.NewInt(req.Amount))))
	if err := validateMsgSend(msg); err != nil {
		_ = s.db.UpdateRequestFailed(dbReq.ID, err.Error())
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Serialize signing/broadcast to avoid sequence reuse under load.
	s.signMutex.Lock()
	defer s.signMutex.Unlock()

	txHash, err := s.buildAndBroadcast(ctx, msg)
	if err != nil {
		if updateErr := s.db.UpdateRequestFailed(dbReq.ID, err.Error()); updateErr != nil {
			log.WithError(updateErr).Error("Failed to update request status")
		}
		return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	if err := s.db.UpdateRequestSuccess(dbReq.ID, txHash); err != nil {
		log.WithError(err).Error("Failed to update request status")
	}

	log.WithFields(log.Fields{
		"tx_hash":   txHash,
		"recipient": req.Recipient,
		"amount":    req.Amount,
	}).Info("Tokens sent successfully")

	return &SendResponse{
		TxHash:    txHash,
		Recipient: req.Recipient,
		Amount:    req.Amount,
	}, nil
}

// GetBalance returns the faucet account balance
func (s *Service) GetBalance() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := s.bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: s.cfg.FaucetAddress,
		Denom:   s.cfg.Denom,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to query balance: %w", err)
	}

	if resp == nil || resp.Balance == nil {
		return 0, fmt.Errorf("balance response empty")
	}

	return resp.Balance.Amount.Int64(), nil
}

// GetNodeStatus returns the blockchain node status
func (s *Service) GetNodeStatus() (*NodeStatus, error) {
	url := fmt.Sprintf("%s/status", s.cfg.NodeRPC)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get node status: status %d, body: %s", resp.StatusCode, string(body))
	}

	var status NodeStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &status, nil
}

// ValidateAddress validates an AURA address
func (s *Service) ValidateAddress(address string) error {
	if address == "" {
		return errors.New("address is required")
	}

	_, err := sdk.AccAddressFromBech32(address)
	return err
}

func (s *Service) buildAndBroadcast(ctx context.Context, msg sdk.Msg) (string, error) {
	accNum, seq, err := s.queryAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to query account: %w", err)
	}

	txBuilder := s.encodingConfig.TxConfig.NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return "", fmt.Errorf("failed to set message: %w", err)
	}

	txBuilder.SetMemo(s.cfg.TransactionMemo)
	txBuilder.SetGasLimit(s.cfg.GasLimit)
	txBuilder.SetFeeAmount(s.feeAmount)

	factory := s.txFactory.
		WithAccountNumber(accNum).
		WithSequence(seq).
		WithGas(s.cfg.GasLimit).
		WithFees(s.feeAmount.String()).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithChainID(s.cfg.ChainID)

	if err := tx.Sign(ctx, factory, s.fromName, txBuilder, true); err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	txBytes, err := s.encodingConfig.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to encode transaction: %w", err)
	}

	rpcClient, err := rpchttp.New(s.cfg.NodeRPC, "/websocket")
	if err != nil {
		return "", fmt.Errorf("failed to create rpc client: %w", err)
	}

	res, err := rpcClient.BroadcastTxSync(ctx, txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	if res.Code != 0 {
		return "", fmt.Errorf("transaction failed with code %d: %s", res.Code, res.Log)
	}

	return hex.EncodeToString(res.Hash), nil
}

func (s *Service) queryAccount(ctx context.Context) (uint64, uint64, error) {
	resp, err := s.authClient.Account(ctx, &authtypes.QueryAccountRequest{
		Address: s.fromAddress.String(),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query account: %w", err)
	}

	var account authtypes.AccountI
	if err := s.encodingConfig.InterfaceRegistry.UnpackAny(resp.Account, &account); err != nil {
		return 0, 0, fmt.Errorf("failed to unpack account: %w", err)
	}

	return account.GetAccountNumber(), account.GetSequence(), nil
}

func initSDKConfig() {
	cfg := sdk.GetConfig()
	if cfg.GetBech32AccountAddrPrefix() == "aura" {
		return
	}

	cfg.SetBech32PrefixForAccount("aura", "aurapub")
	cfg.SetBech32PrefixForValidator("auravaloper", "auravaloperpub")
	cfg.SetBech32PrefixForConsensusNode("auravalcons", "auravalconspub")
	cfg.SetCoinType(sdk.CoinType)
	cfg.SetPurpose(sdk.Purpose)
	cfg.SetFullFundraiserPath(hd.CreateHDPath(sdk.CoinType, 0, 0).String())
	cfg.Seal()
}

func newEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)

	amino := codec.NewLegacyAmino()
	std.RegisterLegacyAminoCodec(amino)
	authtypes.RegisterLegacyAminoCodec(amino)
	banktypes.RegisterLegacyAminoCodec(amino)

	cdc := codec.NewProtoCodec(interfaceRegistry)

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          txConfig,
	}
}

func validateMsgSend(msg *banktypes.MsgSend) error {
	if msg == nil {
		return errors.New("message cannot be nil")
	}

	if len(msg.Amount) == 0 || !msg.Amount.IsAllPositive() {
		return errors.New("amount must be positive")
	}

	return nil
}
