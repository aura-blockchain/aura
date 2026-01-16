package compliance

import (
	"context"
	"strings"
	"testing"
)

// Validation-only tests to ensure client catches bad inputs before network calls.

func TestSubmitKYCValidation(t *testing.T) {
	c := &Client{}

	_, err := c.SubmitKYC(context.Background(), nil)
	if err == nil || !contains(err, "params cannot be nil") {
		t.Fatalf("expected nil params error, got %v", err)
	}

	_, err = c.SubmitKYC(context.Background(), &SubmitKYCParams{})
	if err == nil || !contains(err, "address is required") {
		t.Fatalf("expected address validation, got %v", err)
	}

	_, err = c.SubmitKYC(context.Background(), &SubmitKYCParams{
		Address:       "aura1invalid",
		Provider:      "kyc-inc",
		PIICommitment: []byte{0x01},
		Jurisdiction:  "US",
	})
	if err == nil || !contains(err, "invalid address") {
		t.Fatalf("expected bech32 validation, got %v", err)
	}
}

func TestReportSuspiciousActivityValidation(t *testing.T) {
	c := &Client{}

	_, err := c.ReportSuspiciousActivity(context.Background(), nil)
	if err == nil || !contains(err, "params cannot be nil") {
		t.Fatalf("expected nil params error")
	}

	_, err = c.ReportSuspiciousActivity(context.Background(), &ReportSuspiciousActivityParams{})
	if err == nil || !contains(err, "reporter is required") {
		t.Fatalf("expected reporter validation")
	}

	_, err = c.ReportSuspiciousActivity(context.Background(), &ReportSuspiciousActivityParams{
		Reporter:     "aura1invalid",
		Address:      "aura1invalid",
		ActivityType: "fraud",
	})
	if err == nil || !contains(err, "invalid reporter address") {
		t.Fatalf("expected bech32 validation")
	}
}

// contains is a tiny helper to avoid importing strings in each test
func contains(err error, sub string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), sub)
}
