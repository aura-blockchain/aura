package types

import pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"

// DefaultParams returns default module parameters
func DefaultParams() *pb.ContractRegistryParams {
	return &pb.ContractRegistryParams{
		OpenRegistration:        true,
		MaxContractsPerCreator:  100,
		RequireMetadata:         false,
		RequireSecurityPolicy:   false,
		RequireComplianceConfig: false,
		AuditWarningDays:        365,
		DefaultRateLimit:        1000,
		DefaultMaxGas:           10000000,
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *pb.GenesisState {
	params := DefaultParams()
	return &pb.GenesisState{
		Params:    params,
		Contracts: []*pb.ContractInfo{},
		Metrics:   []*pb.ContractMetrics{},
	}
}

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params *pb.ContractRegistryParams, contracts []*pb.ContractInfo, metrics []*pb.ContractMetrics) *pb.GenesisState {
	return &pb.GenesisState{
		Params:    params,
		Contracts: contracts,
		Metrics:   metrics,
	}
}

// ValidateGenesis performs basic genesis state validation
func ValidateGenesis(gs *pb.GenesisState) error {
	if gs.Params != nil {
		if err := ValidateParams(gs.Params); err != nil {
			return err
		}
	}

	// Validate each contract
	seenAddresses := make(map[string]bool)
	for _, contract := range gs.Contracts {
		if contract.Address == "" {
			return ErrInvalidContractAddress
		}
		if seenAddresses[contract.Address] {
			return ErrContractAlreadyExists
		}
		seenAddresses[contract.Address] = true

		if contract.CodeId == 0 {
			return ErrInvalidCodeID
		}
		if contract.Creator == "" {
			return ErrInvalidRequest
		}
		if contract.Status == pb.ContractStatus_CONTRACT_STATUS_UNSPECIFIED {
			return ErrInvalidRequest
		}
	}

	// Validate metrics match contracts
	for _, metric := range gs.Metrics {
		if metric.ContractAddress == "" {
			return ErrInvalidContractAddress
		}
		if !seenAddresses[metric.ContractAddress] {
			return ErrContractNotFound
		}
	}

	return nil
}

// ValidateParams validates module parameters
func ValidateParams(p *pb.ContractRegistryParams) error {
	if p.MaxContractsPerCreator > 10000 {
		return ErrInvalidParams
	}
	if p.DefaultRateLimit > 10000 {
		return ErrInvalidParams
	}
	if p.DefaultMaxGas > 50000000 {
		return ErrInvalidParams
	}
	return nil
}
