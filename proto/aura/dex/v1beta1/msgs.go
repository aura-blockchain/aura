package v1beta1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSigners returns the signer addresses for MsgCreatePool
// The creator field specifies who is creating the pool and providing initial liquidity
func (m *MsgCreatePool) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signer addresses for MsgAddLiquidity
// The provider field specifies who is adding liquidity to the pool
func (m *MsgAddLiquidity) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(m.Provider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{provider}
}

// GetSigners returns the signer addresses for MsgRemoveLiquidity
// The provider field specifies who is removing liquidity from the pool
func (m *MsgRemoveLiquidity) GetSigners() []sdk.AccAddress {
	provider, err := sdk.AccAddressFromBech32(m.Provider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{provider}
}

// GetSigners returns the signer addresses for MsgSwapExactIn
// The sender field specifies who is executing the swap
func (m *MsgSwapExactIn) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSigners returns the signer addresses for MsgCreateOrder
// The creator field specifies who is creating the P2P swap order
func (m *MsgCreateOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signer addresses for MsgCancelOrder
// The creator field specifies who is canceling the order (must be order owner)
func (m *MsgCancelOrder) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(m.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSigners returns the signer addresses for MsgExecuteSwap
// The initiator field specifies who is executing the matched swap
func (m *MsgExecuteSwap) GetSigners() []sdk.AccAddress {
	initiator, err := sdk.AccAddressFromBech32(m.Initiator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{initiator}
}

// GetSigners returns the signer addresses for MsgCreateHTLC
// The sender field specifies who is creating the Hash Time-Locked Contract
func (m *MsgCreateHTLC) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSigners returns the signer addresses for MsgClaimHTLC
// The recipient field specifies who is claiming the HTLC with the secret
func (m *MsgClaimHTLC) GetSigners() []sdk.AccAddress {
	recipient, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{recipient}
}

// GetSigners returns the signer addresses for MsgRefundHTLC
// The sender field specifies who is refunding the expired HTLC (must be original sender)
func (m *MsgRefundHTLC) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSigners returns the signer addresses for MsgCommitOrder
// The sender field specifies who is committing the order hash (phase 1 of commit-reveal)
func (m *MsgCommitOrder) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// GetSigners returns the signer addresses for MsgRevealOrder
// The sender field specifies who is revealing the committed order (phase 2 of commit-reveal)
func (m *MsgRevealOrder) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}
