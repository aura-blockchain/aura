package types

import "fmt"

// GenesisState defines the initial state of the dataregistry module
type GenesisState struct {
	Params     Params     `json:"params"`
	DataItems  []DataItem `json:"data_items"`
	NextDataID uint64     `json:"next_data_id"`
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:     DefaultParams(),
		DataItems:  []DataItem{},
		NextDataID: 1,
	}
}

// Validate performs basic validation of genesis data
func (gs GenesisState) Validate() error {
	// Validate params
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	// Validate data items
	dataIDs := make(map[string]bool)
	for _, item := range gs.DataItems {
		if item.DataID == "" {
			return fmt.Errorf("data item has empty ID")
		}
		if item.OwnerAddress == "" {
			return fmt.Errorf("data item %s has empty owner address", item.DataID)
		}
		if dataIDs[item.DataID] {
			return fmt.Errorf("duplicate data ID: %s", item.DataID)
		}
		dataIDs[item.DataID] = true
	}

	// Validate next data ID
	if gs.NextDataID == 0 {
		return fmt.Errorf("next data ID must be positive")
	}

	return nil
}
