package v1beta1

import (
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	validAddress  = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validAddress2 = "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"
	validHash     = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
)

// validSignature returns a valid 64-byte signature for testing validator attestations
func validSignature() []byte {
	return make([]byte, 64)
}

// newCoinPtr creates a Coin and returns a pointer for proto-generated types that expect *Coin
// This is necessary because protoc-gen-go generates nullable fields despite gogoproto annotations
func newCoinPtr(denom string, amount int64) *sdk.Coin {
	coin := sdk.NewCoin(denom, sdkmath.NewInt(amount))
	return &coin
}

// newIntString converts an integer amount to string for proto-generated Amount fields
// This is necessary because protoc-gen-go ignores customtype annotations
func newIntString(amount int64) string {
	return sdkmath.NewInt(amount).String()
}

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
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("aura", 1000),
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgLockTokens{
				Sender:      "invalid",
				TargetChain: "paw",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("aura", 1000),
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "empty target chain",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("aura", 1000),
			},
			wantErr: true,
			errMsg:  "target_chain",
		},
		{
			name: "empty recipient",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "",
				Amount:      newCoinPtr("aura", 1000),
			},
			wantErr: true,
			errMsg:  "recipient",
		},
		{
			name: "zero amount",
			msg: &MsgLockTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("aura", 0),
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
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: validSignature(),
			},
			wantErr: false,
		},
		{
			name: "invalid validator",
			msg: &MsgMintTokens{
				Validator:          "invalid",
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: validSignature(),
			},
			wantErr: true,
			errMsg:  "validator",
		},
		{
			name: "empty source chain",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "",
				SourceTxHash:       validHash,
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: validSignature(),
			},
			wantErr: true,
			errMsg:  "source_chain",
		},
		{
			name: "invalid source tx hash",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       "short",
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: validSignature(),
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
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: validSignature(),
			},
			wantErr: true,
			errMsg:  "recipient",
		},
		{
			name: "invalid denom",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "",
				ValidatorSignature: validSignature(),
			},
			wantErr: true,
			errMsg:  "denom",
		},
		{
			name: "signature too small",
			msg: &MsgMintTokens{
				Validator:          validAddress,
				SourceChain:        "paw",
				SourceTxHash:       validHash,
				Recipient:          validAddress2,
				Amount:             newIntString(1000),
				Denom:              "paw.token",
				ValidatorSignature: make([]byte, 32), // Too small
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
			name: "valid message with one signature",
			msg: &MsgUnlockTokens{
				Sender:               validAddress,
				SourceChain:          "paw",
				BurnTxHash:           validHash,
				Amount:               newIntString(1000),
				Denom:                "aura",
				ValidatorSignatures:  [][]byte{validSignature()},
			},
			wantErr: false,
		},
		{
			name: "valid message with multiple signatures",
			msg: &MsgUnlockTokens{
				Sender:               validAddress,
				SourceChain:          "paw",
				BurnTxHash:           validHash,
				Amount:               newIntString(1000),
				Denom:                "aura",
				ValidatorSignatures:  [][]byte{validSignature(), validSignature(), validSignature()},
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgUnlockTokens{
				Sender:               "invalid",
				SourceChain:          "paw",
				BurnTxHash:           validHash,
				Amount:               newIntString(1000),
				Denom:                "aura",
				ValidatorSignatures:  [][]byte{validSignature()},
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "no signatures",
			msg: &MsgUnlockTokens{
				Sender:               validAddress,
				SourceChain:          "paw",
				BurnTxHash:           validHash,
				Amount:               newIntString(1000),
				Denom:                "aura",
				ValidatorSignatures:  [][]byte{},
			},
			wantErr: true,
			errMsg:  "must have at least",
		},
		{
			name: "invalid burn tx hash",
			msg: &MsgUnlockTokens{
				Sender:               validAddress,
				SourceChain:          "paw",
				BurnTxHash:           "short",
				Amount:               newIntString(1000),
				Denom:                "aura",
				ValidatorSignatures:  [][]byte{validSignature()},
			},
			wantErr: true,
			errMsg:  "burn_tx_hash",
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
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("paw.token", 1000),
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgBurnTokens{
				Sender:      "invalid",
				TargetChain: "paw",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("paw.token", 1000),
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "empty target chain",
			msg: &MsgBurnTokens{
				Sender:      validAddress,
				TargetChain: "",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("paw.token", 1000),
			},
			wantErr: true,
			errMsg:  "target_chain",
		},
		{
			name: "zero amount",
			msg: &MsgBurnTokens{
				Sender:      validAddress,
				TargetChain: "paw",
				Recipient:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				Amount:      newCoinPtr("paw.token", 0),
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
				PawAddress:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				XaiAddress:   "",
				PawSignature: validSignature(),
				XaiSignature: nil,
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "valid with XAI address",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "",
				XaiAddress:   "xai1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				PawSignature: nil,
				XaiSignature: validSignature(),
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "valid with both addresses",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				XaiAddress:   "xai1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				PawSignature: validSignature(),
				XaiSignature: validSignature(),
				Signer:       validAddress,
			},
			wantErr: false,
		},
		{
			name: "invalid AURA address",
			msg: &MsgLinkAddress{
				AuraAddress:  "invalid",
				PawAddress:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				PawSignature: validSignature(),
				Signer:       "invalid",
			},
			wantErr: true,
			errMsg:  "aura_address",
		},
		{
			name: "no addresses to link",
			msg: &MsgLinkAddress{
				AuraAddress: validAddress,
				PawAddress:  "",
				XaiAddress:  "",
				Signer:      validAddress,
			},
			wantErr: true,
			errMsg:  "at least one of paw_address or xai_address must be provided",
		},
		{
			name: "signer not in addresses",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				PawSignature: validSignature(),
				Signer:       validAddress2,
			},
			wantErr: true,
			errMsg:  "signer must be one of the addresses being linked",
		},
		{
			name: "PAW address without signature",
			msg: &MsgLinkAddress{
				AuraAddress:  validAddress,
				PawAddress:   "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				PawSignature: nil,
				Signer:       validAddress,
			},
			wantErr: true,
			errMsg:  "paw_signature",
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
				InputCoin:       newCoinPtr("aura", 1000),
				TargetChain:     "paw",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
				Recipient:       "",
				MaxSlippageBps:  500, // 5%
			},
			wantErr: false,
		},
		{
			name: "valid with recipient",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       newCoinPtr("aura", 1000),
				TargetChain:     "paw",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
				Recipient:       "paw1qypqxpq9qcrsszg2pvxq6rs0zqg3yycdefghij",
				MaxSlippageBps:  500,
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgCrossChainSwap{
				Sender:          "invalid",
				SourceChain:     "aura",
				InputCoin:       newCoinPtr("aura", 1000),
				TargetChain:     "paw",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
				MaxSlippageBps:  500,
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "same source and target chain",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       newCoinPtr("aura", 1000),
				TargetChain:     "aura",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
				MaxSlippageBps:  500,
			},
			wantErr: true,
			errMsg:  "source_chain and target_chain must be different",
		},
		{
			name: "zero input amount",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       newCoinPtr("aura", 0),
				TargetChain:     "paw",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
				MaxSlippageBps:  500,
			},
			wantErr: true,
			errMsg:  "input_coin",
		},
		{
			name: "slippage too high",
			msg: &MsgCrossChainSwap{
				Sender:          validAddress,
				SourceChain:     "aura",
				InputCoin:       newCoinPtr("aura", 1000),
				TargetChain:     "paw",
				TargetDenom:     "paw",
				MinTargetAmount: newIntString(900),
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
			name: "valid message",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "completed",
			},
			wantErr: false,
		},
		{
			name: "invalid relayer",
			msg: &MsgRelayTransfer{
				Relayer:      "invalid",
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       "completed",
			},
			wantErr: true,
			errMsg:  "relayer",
		},
		{
			name: "empty transfer ID",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "",
				TargetTxHash: validHash,
				Status:       "completed",
			},
			wantErr: true,
			errMsg:  "transfer_id",
		},
		{
			name: "invalid target tx hash",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: "short",
				Status:       "completed",
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
				Status:       "",
			},
			wantErr: true,
			errMsg:  "status",
		},
		{
			name: "status too long",
			msg: &MsgRelayTransfer{
				Relayer:      validAddress,
				TransferId:   "transfer-123",
				TargetTxHash: validHash,
				Status:       strings.Repeat("a", 65),
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
