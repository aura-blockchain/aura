package bridge

import (
	"context"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestLockTokensValidation(t *testing.T) {
	c := &Client{}

	_, err := c.LockTokens(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error, got %v", err)
	}

	_, err = c.LockTokens(context.Background(), &LockTokensParams{})
	if err == nil || !strings.Contains(err.Error(), "sender is required") {
		t.Fatalf("expected sender validation error")
	}

	_, err = c.LockTokens(context.Background(), &LockTokensParams{
		Sender:      "aura1invalid",
		TargetChain: "",
		Recipient:   "",
		Amount:      sdk.NewCoin("uaura", math.ZeroInt()),
	})
	if err == nil || !strings.Contains(err.Error(), "target chain is required") {
		t.Fatalf("expected target chain validation")
	}

	_, err = c.LockTokens(context.Background(), &LockTokensParams{
		Sender:      "aura1invalid",
		TargetChain: "paw",
		Recipient:   "",
		Amount:      sdk.NewCoin("uaura", math.ZeroInt()),
	})
	if err == nil || !strings.Contains(err.Error(), "recipient is required") {
		t.Fatalf("expected recipient validation")
	}

	_, err = c.LockTokens(context.Background(), &LockTokensParams{
		Sender:      "aura1invalid",
		TargetChain: "paw",
		Recipient:   "paw1receiver",
		Amount:      sdk.NewCoin("uaura", math.ZeroInt()),
	})
	if err == nil || !strings.Contains(err.Error(), "amount must be positive") {
		t.Fatalf("expected positive amount validation")
	}
}
