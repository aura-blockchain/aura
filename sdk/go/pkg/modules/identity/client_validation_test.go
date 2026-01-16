package identity

import (
	"context"
	"strings"
	"testing"
)

// These tests focus on client-side validation paths that short-circuit before any network calls.

func TestRequestIdentityChangeValidation(t *testing.T) {
	c := &Client{} // auraClient not needed; validation should fail before broadcast

	_, err := c.RequestIdentityChange(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error, got %v", err)
	}

	_, err = c.RequestIdentityChange(context.Background(), &RequestIdentityChangeParams{
		Requester: "",
		TargetDID: "",
	})
	if err == nil || !strings.Contains(err.Error(), "requester is required") {
		t.Fatalf("expected requester validation error, got %v", err)
	}

	_, err = c.RequestIdentityChange(context.Background(), &RequestIdentityChangeParams{
		Requester: "notabech32",
		TargetDID: "did:aura:123",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid requester address") {
		t.Fatalf("expected bech32 validation error, got %v", err)
	}
}

func TestCreateRoleValidation(t *testing.T) {
	c := &Client{}

	_, err := c.CreateRole(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error")
	}

	_, err = c.CreateRole(context.Background(), &CreateRoleParams{
		Creator:     "",
		RoleName:    "",
		Permissions: []string{},
	})
	if err == nil || !strings.Contains(err.Error(), "creator is required") {
		t.Fatalf("expected creator validation error")
	}

	longName := strings.Repeat("x", 257)
	_, err = c.CreateRole(context.Background(), &CreateRoleParams{
		Creator:     "aura1invalid", // invalid bech32 to stop before broadcast
		RoleName:    longName,
		Permissions: []string{"read"},
	})
	if err == nil || !strings.Contains(err.Error(), "role name exceeds") {
		t.Fatalf("expected role length validation")
	}

	_, err = c.CreateRole(context.Background(), &CreateRoleParams{
		Creator:     "aura1invalid",
		RoleName:    "ok",
		Permissions: []string{""},
	})
	if err == nil || !strings.Contains(err.Error(), "permission at index 0 is empty") {
		t.Fatalf("expected permission validation")
	}
}

func TestAssignRoleValidation(t *testing.T) {
	c := &Client{}

	_, err := c.AssignRole(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error")
	}

	_, err = c.AssignRole(context.Background(), &AssignRoleParams{
		Assigner: "",
		Address:  "",
		RoleName: "",
	})
	if err == nil || !strings.Contains(err.Error(), "assigner is required") {
		t.Fatalf("expected assigner validation")
	}

	_, err = c.AssignRole(context.Background(), &AssignRoleParams{
		Assigner: "aura1invalid",
		Address:  "aura1invalid",
		RoleName: "role",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid assigner address") {
		t.Fatalf("expected bech32 validation")
	}
}
