package types

// GenesisState defines the monitoring module's genesis state
type GenesisState struct {
	Params Params `json:"params"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	return ValidateParams(gs.Params)
}
