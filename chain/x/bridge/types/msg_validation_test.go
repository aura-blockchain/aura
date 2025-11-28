package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	validAddress   = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validHash      = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	validSignature = "3045022100abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456789002200fedcba0987654321fedcba0987654321fedcba0987654321fedcba09876543"
)

func TestMsgLockTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgLockTokens
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: false,
		},
		{
			name: "invalid sender address",
			msg: &MsgLockTokens{
				Sender:      "invalid",
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "empty sender",
			msg: &MsgLockTokens{
				Sender:      "",
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "too short target chain",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "x", // too short - min is 2 chars
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: true,
			errMsg:  "target_chain",
		},
		{
			name: "valid xai target chain",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "xai",
				Recipient:   "xai1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: false,
		},
		{
			name: "empty recipient",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: true,
			errMsg:  "recipient",
		},
		{
			name: "nil amount",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      nil,
			},
			wantErr: true,
			errMsg:  "amount cannot be nil",
		},
		{
			name: "zero amount",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(0)},
			},
			wantErr: true,
			errMsg:  "amount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgMintTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgMintTokens
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: false,
		},
		{
			name: "invalid validator address",
			msg: &MsgMintTokens{
				Validator:          "invalid",
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "validator",
		},
		{
			name: "too short source chain",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "x", // too short - min is 2 chars
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "source_chain",
		},
		{
			name: "invalid source tx hash",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       "invalid",
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "source_tx_hash",
		},
		{
			name: "invalid recipient",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          "invalid",
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "recipient",
		},
		{
			name: "zero amount",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "0",
				Denom:              "uatom",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "amount",
		},
		{
			name: "invalid denom",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "1invalid",
				ValidatorSignature: []byte(validSignature),
			},
			wantErr: true,
			errMsg:  "denom",
		},
		{
			name: "signature too short",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress,
				Amount:             "1000",
				Denom:              "uatom",
				ValidatorSignature: []byte("short"),
			},
			wantErr: true,
			errMsg:  "validator_signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUnlockTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgUnlockTokens
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgUnlockTokens{
				Sender:              validAddress,
				SourceChain:         "paw",
				BurnTxHash:          validHash,
				Amount:              "1000",
				Denom:               "uatom",
				ValidatorSignatures: [][]byte{[]byte(validSignature)},
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgUnlockTokens{
				Sender:              "invalid",
				SourceChain:         "paw",
				BurnTxHash:          validHash,
				Amount:              "1000",
				Denom:               "uatom",
				ValidatorSignatures: [][]byte{[]byte(validSignature)},
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "invalid burn tx hash",
			msg: &MsgUnlockTokens{
				Sender:              validAddress,
				SourceChain:         "paw",
				BurnTxHash:          "invalid",
				Amount:              "1000",
				Denom:               "uatom",
				ValidatorSignatures: [][]byte{[]byte(validSignature)},
			},
			wantErr: true,
			errMsg:  "burn_tx_hash",
		},
		{
			name: "empty validator signatures",
			msg: &MsgUnlockTokens{
				Sender:              validAddress,
				SourceChain:         "paw",
				BurnTxHash:          validHash,
				Amount:              "1000",
				Denom:               "uatom",
				ValidatorSignatures: [][]byte{},
			},
			wantErr: true,
			errMsg:  "must have at least 1 signature",
		},
		{
			name: "invalid signature in slice",
			msg: &MsgUnlockTokens{
				Sender:              validAddress,
				SourceChain:         "paw",
				BurnTxHash:          validHash,
				Amount:              "1000",
				Denom:               "uatom",
				ValidatorSignatures: [][]byte{[]byte("short")},
			},
			wantErr: true,
			errMsg:  "validator_signatures[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgBurnTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgBurnTokens
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgBurnTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "paw.token", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgBurnTokens{
				Sender:      "invalid",
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      &sdk.Coin{Denom: "paw.token", Amount: sdkmath.NewInt(1000)},
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "nil amount",
			msg: &MsgBurnTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1abc123",
				Amount:      nil,
			},
			wantErr: true,
			errMsg:  "amount cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgLinkAddress_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgLinkAddress
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with PAW address",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "paw1abc123",
				PawSignature: []byte(validSignature),
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "valid with XAI address",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				XaiAddress:   "xai1abc123",
				XaiSignature: []byte(validSignature),
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "valid with both addresses",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "paw1abc123",
				PawSignature: []byte(validSignature),
				XaiAddress:   "xai1abc123",
				XaiSignature: []byte(validSignature),
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "invalid - no addresses provided",
			msg: &MsgLinkAddress{
				AuraAddress: validAddress,
				Signer:      validAddress,
			},
			wantErr: true,
			errMsg:  "at least one of paw_address or xai_address must be provided",
		},
		{
			name: "invalid AURA address",
			msg: &MsgLinkAddress{
				AuraAddress:  "invalid",
				PawAddress:   "paw1abc123",
				PawSignature: []byte(validSignature),
				Signer:       validAddress,
			},
			wantErr: true,
			errMsg:  "aura_address",
		},
		{
			name: "PAW address without signature",
			msg: &MsgLinkAddress{
				AuraAddress: validAddress,
				PawAddress:  "paw1abc123",
				Signer:      validAddress,
			},
			wantErr: true,
			errMsg:  "paw_signature",
		},
		{
			name: "XAI address without signature",
			msg: &MsgLinkAddress{
				AuraAddress: validAddress,
				XaiAddress:  "xai1abc123",
				Signer:      validAddress,
			},
			wantErr: true,
			errMsg:  "xai_signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCrossChainSwap_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCrossChainSwap
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				TargetChain:     "paw",
				TargetDenom:     "paw.token",
				MinTargetAmount: "900",
				MaxSlippageBps:  500, // 5%
			},
			wantErr: false,
		},
		{
			name: "valid with recipient",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				TargetChain:     "paw",
				TargetDenom:     "paw.token",
				MinTargetAmount: "900",
				Recipient:       validAddress,
				MaxSlippageBps:  500,
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgCrossChainSwap{
				Sender:          "invalid",
				SourceChain:     "aura",
				InputCoin:       &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				TargetChain:     "paw",
				TargetDenom:     "paw.token",
				MinTargetAmount: "900",
				MaxSlippageBps:  500,
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "invalid input coin",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(0)},
				TargetChain:     "paw",
				TargetDenom:     "paw.token",
				MinTargetAmount: "900",
				MaxSlippageBps:  500,
			},
			wantErr: true,
			errMsg:  "input_coin",
		},
		{
			name: "invalid slippage bps",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				TargetChain:     "paw",
				TargetDenom:     "paw.token",
				MinTargetAmount: "900",
				MaxSlippageBps:  10001, // > 100%
			},
			wantErr: true,
			errMsg:  "max_slippage_bps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgRelayTransfer_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRelayTransfer
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message - pending",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "pending",
			},
			wantErr: false,
		},
		{
			name: "valid message - confirmed",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "confirmed",
			},
			wantErr: false,
		},
		{
			name: "valid message - completed",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "completed",
			},
			wantErr: false,
		},
		{
			name: "valid message - failed",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "failed",
			},
			wantErr: false,
		},
		{
			name: "invalid relayer",
			msg: &MsgRelayTransfer{
				Relayer:      "invalid",
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "relayer",
		},
		{
			name: "invalid transfer ID",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "",
				TargetTxHash: validHash,
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "transfer_id",
		},
		{
			name: "invalid target tx hash",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: "invalid",
				Status:       "pending",
			},
			wantErr: true,
			errMsg:  "target_tx_hash",
		},
		{
			name: "empty status",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "", // empty status fails validation
			},
			wantErr: true,
			errMsg:  "status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Note: GetSigners() tests removed - proto-generated types don't include this method
// In Cosmos SDK v0.50+, GetSigners is implemented via amino annotations in proto files
