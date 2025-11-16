package types

// SharedIdentity represents a linked identity across multiple chains
type SharedIdentity struct {
	AuraAddress     string            `json:"aura_address"`
	LinkedAddresses map[string]string `json:"linked_addresses"` // chain -> address
	AuraIrScore     uint64            `json:"aura_ir_score"`
	VerifiedAura    bool              `json:"verified_aura"`
	VerifiedPaw     bool              `json:"verified_paw"`
	VerifiedXai     bool              `json:"verified_xai"`
}
