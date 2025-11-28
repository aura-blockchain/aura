package v1beta1

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	validAddress = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validHash    = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

func TestMsgCreatePool_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreatePool
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCreatePool{
				Creator: validAddress,
				DenomA:  "uatom",
				DenomB:  "uosmo",
				AmountA: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB: &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: false,
		},
		{
			name: "invalid creator address",
			msg: &MsgCreatePool{
				Creator: "invalid",
				DenomA:  "uatom",
				DenomB:  "uosmo",
				AmountA: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB: &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "empty creator",
			msg: &MsgCreatePool{
				Creator: "",
				DenomA:  "uatom",
				DenomB:  "uosmo",
				AmountA: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB: &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "same denoms",
			msg: &MsgCreatePool{
				Creator: validAddress,
				DenomA:  "uatom",
				DenomB:  "uatom",
				AmountA: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "must be different",
		},
		{
			name: "zero amount A",
			msg: &MsgCreatePool{
				Creator: validAddress,
				DenomA:  "uatom",
				DenomB:  "uosmo",
				AmountA: &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(0)},
				AmountB: &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "amount_a",
		},
		{
			name: "nil amount A",
			msg: &MsgCreatePool{
				Creator: validAddress,
				DenomA:  "uatom",
				DenomB:  "uosmo",
				AmountA: nil,
				AmountB: &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "amount_a cannot be nil",
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

func TestMsgAddLiquidity_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgAddLiquidity
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgAddLiquidity{
				Provider: validAddress,
				PoolId:   "pool-1",
				AmountA:  &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB:  &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			msg: &MsgAddLiquidity{
				Provider: "invalid",
				PoolId:   "pool-1",
				AmountA:  &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB:  &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "provider",
		},
		{
			name: "empty pool ID",
			msg: &MsgAddLiquidity{
				Provider: validAddress,
				PoolId:   "",
				AmountA:  &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB:  &sdk.Coin{Denom: "uosmo", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "pool_id",
		},
		{
			name: "same denoms",
			msg: &MsgAddLiquidity{
				Provider: validAddress,
				PoolId:   "pool-1",
				AmountA:  &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				AmountB:  &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(2000)},
			},
			wantErr: true,
			errMsg:  "must have different denoms",
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

func TestMsgRemoveLiquidity_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRemoveLiquidity
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgRemoveLiquidity{
				Provider: validAddress,
				PoolId:   "pool-1",
				LpTokens: "1000",
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			msg: &MsgRemoveLiquidity{
				Provider: "invalid",
				PoolId:   "pool-1",
				LpTokens: "1000",
			},
			wantErr: true,
			errMsg:  "provider",
		},
		{
			name: "zero LP tokens",
			msg: &MsgRemoveLiquidity{
				Provider: validAddress,
				PoolId:   "pool-1",
				LpTokens: "0",
			},
			wantErr: true,
			errMsg:  "lp_tokens",
		},
		{
			name: "negative LP tokens",
			msg: &MsgRemoveLiquidity{
				Provider: validAddress,
				PoolId:   "pool-1",
				LpTokens: "-100",
			},
			wantErr: true,
			errMsg:  "lp_tokens",
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

func TestMsgSwapExactIn_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgSwapExactIn
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgSwapExactIn{
				Sender:         validAddress,
				PoolId:         "pool-1",
				CoinIn:         &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				MinAmountOut:   "900",
				MaxSlippageBps: 500, // 5%
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgSwapExactIn{
				Sender:         "invalid",
				PoolId:         "pool-1",
				CoinIn:         &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				MinAmountOut:   "900",
				MaxSlippageBps: 500,
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "slippage too high",
			msg: &MsgSwapExactIn{
				Sender:         validAddress,
				PoolId:         "pool-1",
				CoinIn:         &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				MinAmountOut:   "900",
				MaxSlippageBps: 10001, // > 100%
			},
			wantErr: true,
			errMsg:  "max_slippage_bps",
		},
		{
			name: "zero min amount out",
			msg: &MsgSwapExactIn{
				Sender:         validAddress,
				PoolId:         "pool-1",
				CoinIn:         &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				MinAmountOut:   "0",
				MaxSlippageBps: 500,
			},
			wantErr: true,
			errMsg:  "min_amount_out",
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

func TestMsgCreateOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateOrder
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid buy order",
			msg: &MsgCreateOrder{
				Creator:     validAddress,
				OrderType:   SwapOrderType_BUY,
				AuraAmount:  "1000",
				OtherCoin:   "usdt",
				OtherAmount: "5000",
			},
			wantErr: false,
		},
		{
			name: "valid sell order",
			msg: &MsgCreateOrder{
				Creator:     validAddress,
				OrderType:   SwapOrderType_SELL,
				AuraAmount:  "1000",
				OtherCoin:   "usdt",
				OtherAmount: "5000",
			},
			wantErr: false,
		},
		{
			name: "invalid creator",
			msg: &MsgCreateOrder{
				Creator:     "invalid",
				OrderType:   SwapOrderType_BUY,
				AuraAmount:  "1000",
				OtherCoin:   "usdt",
				OtherAmount: "5000",
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "zero AURA amount",
			msg: &MsgCreateOrder{
				Creator:     validAddress,
				OrderType:   SwapOrderType_BUY,
				AuraAmount:  "0",
				OtherCoin:   "usdt",
				OtherAmount: "5000",
			},
			wantErr: true,
			errMsg:  "aura_amount",
		},
		{
			name: "invalid other coin denom",
			msg: &MsgCreateOrder{
				Creator:     validAddress,
				OrderType:   SwapOrderType_BUY,
				AuraAmount:  "1000",
				OtherCoin:   "1invalid",
				OtherAmount: "5000",
			},
			wantErr: true,
			errMsg:  "other_coin",
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

func TestMsgCancelOrder_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCancelOrder
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCancelOrder{
				Creator: validAddress,
				OrderId: "order-123",
			},
			wantErr: false,
		},
		{
			name: "invalid creator",
			msg: &MsgCancelOrder{
				Creator: "invalid",
				OrderId: "order-123",
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "empty order ID",
			msg: &MsgCancelOrder{
				Creator: validAddress,
				OrderId: "",
			},
			wantErr: true,
			errMsg:  "order_id",
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

func TestMsgCreateHTLC_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateHTLC
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCreateHTLC{
				Sender:           validAddress,
				Recipient:        "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
				Amount:           &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				SecretHash:       validHash,
				TimelockDuration: 3600, // 1 hour
			},
			wantErr: false,
		},
		{
			name: "sender equals recipient",
			msg: &MsgCreateHTLC{
				Sender:           validAddress,
				Recipient:        validAddress,
				Amount:           &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				SecretHash:       validHash,
				TimelockDuration: 3600,
			},
			wantErr: true,
			errMsg:  "must be different",
		},
		{
			name: "timelock too short",
			msg: &MsgCreateHTLC{
				Sender:           validAddress,
				Recipient:        "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
				Amount:           &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				SecretHash:       validHash,
				TimelockDuration: 60, // < 1 hour
			},
			wantErr: true,
			errMsg:  "timelock_duration",
		},
		{
			name: "invalid secret hash",
			msg: &MsgCreateHTLC{
				Sender:           validAddress,
				Recipient:        "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
				Amount:           &sdk.Coin{Denom: "uatom", Amount: sdkmath.NewInt(1000)},
				SecretHash:       "short",
				TimelockDuration: 3600,
			},
			wantErr: true,
			errMsg:  "secret_hash",
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

func TestMsgClaimHTLC_ValidateBasic(t *testing.T) {
	validSecret := "thisisavalidsecretthatisat32byteslongenoughforvalidation"
	tests := []struct {
		name    string
		msg     *MsgClaimHTLC
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgClaimHTLC{
				Recipient: validAddress,
				HtlcId:    "htlc-123",
				Secret:    validSecret,
			},
			wantErr: false,
		},
		{
			name: "invalid recipient",
			msg: &MsgClaimHTLC{
				Recipient: "invalid",
				HtlcId:    "htlc-123",
				Secret:    validSecret,
			},
			wantErr: true,
			errMsg:  "recipient",
		},
		{
			name: "empty HTLC ID",
			msg: &MsgClaimHTLC{
				Recipient: validAddress,
				HtlcId:    "",
				Secret:    validSecret,
			},
			wantErr: true,
			errMsg:  "htlc_id",
		},
		{
			name: "short secret",
			msg: &MsgClaimHTLC{
				Recipient: validAddress,
				HtlcId:    "htlc-123",
				Secret:    "short",
			},
			wantErr: true,
			errMsg:  "secret",
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

func TestMsgRefundHTLC_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRefundHTLC
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgRefundHTLC{
				Sender: validAddress,
				HtlcId: "htlc-123",
			},
			wantErr: false,
		},
		{
			name: "invalid sender",
			msg: &MsgRefundHTLC{
				Sender: "invalid",
				HtlcId: "htlc-123",
			},
			wantErr: true,
			errMsg:  "sender",
		},
		{
			name: "empty HTLC ID",
			msg: &MsgRefundHTLC{
				Sender: validAddress,
				HtlcId: "",
			},
			wantErr: true,
			errMsg:  "htlc_id",
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
