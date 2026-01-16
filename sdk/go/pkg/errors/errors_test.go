package errors

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionErrors(t *testing.T) {
	err := NewConnectionError("down")
	assert.Equal(t, "CONNECTION_ERROR", err.Code)
	assert.True(t, stderrors.Is(err, ErrConnection))

	notConnected := NewNotConnectedError("broadcast")
	assert.Contains(t, notConnected.Message, "broadcast")
	assert.True(t, stderrors.Is(notConnected, ErrNotConnected))

	endpointErr := NewEndpointError("http://rpc", "timeout")
	assert.Equal(t, "ENDPOINT_ERROR", endpointErr.Code)
	assert.Equal(t, "http://rpc", endpointErr.Endpoint)
}

func TestWalletErrors(t *testing.T) {
	err := NewInvalidMnemonicError()
	assert.True(t, stderrors.Is(err, ErrInvalidMnemonic))
	assert.Equal(t, "INVALID_MNEMONIC", err.Code)

	locked := NewWalletNotInitializedError("sign")
	assert.Contains(t, locked.Message, "sign")
}

func TestTransactionErrors(t *testing.T) {
	timeout := NewTxTimeoutError("0xabc", 5000)
	assert.Equal(t, "TX_TIMEOUT", timeout.Code)
	assert.Equal(t, "0xabc", timeout.TxHash)

	insuff := NewInsufficientFundsError("10", "5", "uaura")
	assert.True(t, stderrors.Is(insuff, ErrInsufficientFunds))
	assert.Contains(t, insuff.Message, "insufficient funds")
}

func TestQueryAndValidationErrors(t *testing.T) {
	notFound := NewNotFoundError("vc", "123")
	assert.Equal(t, "NOT_FOUND", notFound.Code)
	assert.True(t, stderrors.Is(notFound, ErrNotFound))

	invalidAddr := NewInvalidAddressError("bad", "aura")
	assert.Equal(t, "INVALID_ADDRESS", invalidAddr.Code)
	assert.Equal(t, "address", invalidAddr.Field)
}

func TestIdentityAndDEXErrors(t *testing.T) {
	didErr := NewDIDNotFoundError("did:aura:abc")
	assert.True(t, stderrors.Is(didErr, ErrDIDNotFound))
	assert.Equal(t, "did:aura:abc", didErr.DID)

	slippage := NewSlippageExceededError("1.0", "1.2", "0.5")
	assert.True(t, stderrors.Is(slippage, ErrSlippageExceeded))
	assert.Contains(t, slippage.Message, "slippage exceeded")
}

func TestWrapAndIsAuraError(t *testing.T) {
	base := stderrors.New("boom")
	wrapped := Wrap(base, "context")
	assert.True(t, IsAuraError(wrapped))
	assert.Contains(t, wrapped.Message, "context")

	// Wrapping an AuraError should preserve aura classification
	aura := NewWalletError("problem")
	again := Wrap(aura, "ignored")
	assert.True(t, IsAuraError(again))
	assert.Contains(t, again.Message, "problem")
}
