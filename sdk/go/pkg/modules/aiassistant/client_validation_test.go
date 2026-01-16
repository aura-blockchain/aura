package aiassistant

import (
	"context"
	"strings"
	"testing"
)

func TestRegisterAssistantValidation(t *testing.T) {
	c := &Client{}

	_, err := c.RegisterAssistant(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error")
	}

	_, err = c.RegisterAssistant(context.Background(), &RegisterAssistantParams{})
	if err == nil || !strings.Contains(err.Error(), "assistant address is required") {
		t.Fatalf("expected assistant address validation")
	}

	_, err = c.RegisterAssistant(context.Background(), &RegisterAssistantParams{
		AssistantAddress: "aura1invalid",
		OwnerAddress:     "",
		Locales:          []string{},
	})
	if err == nil || !strings.Contains(err.Error(), "owner address is required") {
		t.Fatalf("expected owner address validation")
	}
}

func TestUpdateLocalesValidation(t *testing.T) {
	c := &Client{}

	_, err := c.UpdateLocales(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "params cannot be nil") {
		t.Fatalf("expected nil params error")
	}

	_, err = c.UpdateLocales(context.Background(), &UpdateLocalesParams{})
	if err == nil || !strings.Contains(err.Error(), "assistant address is required") {
		t.Fatalf("expected assistant validation")
	}

	_, err = c.UpdateLocales(context.Background(), &UpdateLocalesParams{
		AssistantAddress: "aura1invalid",
		Locales:          []string{},
	})
	if err == nil || !strings.Contains(err.Error(), "at least one locale is required") {
		t.Fatalf("expected locale validation")
	}
}
