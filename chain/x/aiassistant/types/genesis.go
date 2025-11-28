package types

import (
	"fmt"
)

func DefaultGenesis() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Assistants: []*Assistant{},
		Params:     &params,
	}
}

func ValidateGenesis(gs GenesisState) error {
	if gs.Params == nil {
		return fmt.Errorf("params cannot be nil")
	}
	if err := ValidateParams(*gs.Params); err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, asst := range gs.Assistants {
		if asst == nil {
			return fmt.Errorf("assistant entry cannot be nil")
		}
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

func ValidateAssistant(asst *Assistant) error {
	if asst.AssistantAddress == "" {
		return fmt.Errorf("assistant_address cannot be empty")
	}
	if asst.OwnerAddress == "" {
		return fmt.Errorf("owner_address cannot be empty")
	}
	if asst.Stake == nil {
		return fmt.Errorf("stake cannot be nil")
	}
	if asst.SponsorshipBalance == nil {
		return fmt.Errorf("sponsorship_balance cannot be nil")
	}
	if len(asst.Locales) == 0 {
		return fmt.Errorf("locales cannot be empty")
	}
	return nil
}
