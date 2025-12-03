package v1beta1

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestGetSigners_MsgCreatePool verifies that MsgCreatePool correctly implements GetSigners
func TestGetSigners_MsgCreatePool(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgCreatePool{
		Creator: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgAddLiquidity verifies that MsgAddLiquidity correctly implements GetSigners
func TestGetSigners_MsgAddLiquidity(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgAddLiquidity{
		Provider: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgRemoveLiquidity verifies that MsgRemoveLiquidity correctly implements GetSigners
func TestGetSigners_MsgRemoveLiquidity(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgRemoveLiquidity{
		Provider: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgSwapExactIn verifies that MsgSwapExactIn correctly implements GetSigners
func TestGetSigners_MsgSwapExactIn(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgSwapExactIn{
		Sender: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgCreateOrder verifies that MsgCreateOrder correctly implements GetSigners
func TestGetSigners_MsgCreateOrder(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgCreateOrder{
		Creator: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgCancelOrder verifies that MsgCancelOrder correctly implements GetSigners
func TestGetSigners_MsgCancelOrder(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgCancelOrder{
		Creator: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgExecuteSwap verifies that MsgExecuteSwap correctly implements GetSigners
func TestGetSigners_MsgExecuteSwap(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgExecuteSwap{
		Initiator: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgCreateHTLC verifies that MsgCreateHTLC correctly implements GetSigners
func TestGetSigners_MsgCreateHTLC(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgCreateHTLC{
		Sender: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgClaimHTLC verifies that MsgClaimHTLC correctly implements GetSigners
func TestGetSigners_MsgClaimHTLC(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgClaimHTLC{
		Recipient: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgRefundHTLC verifies that MsgRefundHTLC correctly implements GetSigners
func TestGetSigners_MsgRefundHTLC(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgRefundHTLC{
		Sender: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgCommitOrder verifies that MsgCommitOrder correctly implements GetSigners
func TestGetSigners_MsgCommitOrder(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgCommitOrder{
		Sender: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_MsgRevealOrder verifies that MsgRevealOrder correctly implements GetSigners
func TestGetSigners_MsgRevealOrder(t *testing.T) {
	addr := sdk.AccAddress("test_address_12345")
	msg := &MsgRevealOrder{
		Sender: addr.String(),
	}

	signers := msg.GetSigners()
	require.Len(t, signers, 1)
	require.Equal(t, addr, signers[0])
}

// TestGetSigners_InvalidAddress verifies panic on invalid bech32 address
func TestGetSigners_InvalidAddress(t *testing.T) {
	msg := &MsgCreatePool{
		Creator: "invalid_address",
	}

	require.Panics(t, func() {
		msg.GetSigners()
	})
}

// TestGetSigners_VerifyInterface verifies all messages implement sdk.Msg interface
func TestGetSigners_VerifyInterface(t *testing.T) {
	testCases := []struct {
		name string
		msg  interface{}
	}{
		{"MsgCreatePool", &MsgCreatePool{}},
		{"MsgAddLiquidity", &MsgAddLiquidity{}},
		{"MsgRemoveLiquidity", &MsgRemoveLiquidity{}},
		{"MsgSwapExactIn", &MsgSwapExactIn{}},
		{"MsgCreateOrder", &MsgCreateOrder{}},
		{"MsgCancelOrder", &MsgCancelOrder{}},
		{"MsgExecuteSwap", &MsgExecuteSwap{}},
		{"MsgCreateHTLC", &MsgCreateHTLC{}},
		{"MsgClaimHTLC", &MsgClaimHTLC{}},
		{"MsgRefundHTLC", &MsgRefundHTLC{}},
		{"MsgCommitOrder", &MsgCommitOrder{}},
		{"MsgRevealOrder", &MsgRevealOrder{}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify the message implements the GetSigners interface
			_, ok := tc.msg.(interface{ GetSigners() []sdk.AccAddress })
			require.True(t, ok, "%s must implement GetSigners()", tc.name)
		})
	}
}
