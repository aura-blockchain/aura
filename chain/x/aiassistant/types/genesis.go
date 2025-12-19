package types

import (
	"fmt"
)

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Assistants: []Assistant{},
		Params:     DefaultParams(),
	}
}

func ValidateGenesis(gs GenesisState) error {
	if err := ValidateParams(gs.Params); err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, asst := range gs.Assistants {
		if _, exists := seen[asst.AssistantAddress]; exists {
			return fmt.Errorf("duplicate assistant %s", asst.AssistantAddress)
		}
		seen[asst.AssistantAddress] = struct{}{}
		if err := ValidateAssistant(asst); err != nil {
			return err
		}
	}
	return nil
}

func ValidateAssistant(asst Assistant) error {
	if asst.AssistantAddress == "" {
		return fmt.Errorf("assistant_address cannot be empty")
	}
	if asst.OwnerAddress == "" {
		return fmt.Errorf("owner_address cannot be empty")
	}
	if err := validateBalance(asst.Stake, true); err != nil {
		return fmt.Errorf("stake: %w", err)
	}
	if err := validateBalance(asst.SponsorshipBalance, false); err != nil {
		return fmt.Errorf("sponsorship_balance: %w", err)
	}
	if len(asst.Locales) == 0 {
		return fmt.Errorf("locales cannot be empty")
	}
	return nil
}
