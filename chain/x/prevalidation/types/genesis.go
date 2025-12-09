package types

import (
	pb "github.com/aequitas/aura/proto/aura/prevalidation/v1beta1"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *pb.GenesisState {
	return &pb.GenesisState{
		Params:                   *DefaultParams(),
		PreValidatedTransactions: []*pb.PreValidatedTransaction{},
		Templates:                []*pb.ValidationTemplate{},
		Metrics:                  &pb.PreValidationMetrics{},
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *pb.GenesisState) error {
	if err := ValidateParams(&data.Params); err != nil {
		return err
	}

	// Validate pre-validated transactions
	for _, tx := range data.PreValidatedTransactions {
		if tx == nil {
			continue // Skip nil entries
		}
		if tx.Id == "" {
			return ErrInvalidTransaction
		}
		if tx.TxType == pb.TransactionType_TX_TYPE_UNSPECIFIED {
			return ErrInvalidTransactionType
		}
		if tx.Status == pb.ValidationStatus_VALIDATION_STATUS_UNSPECIFIED {
			return ErrInvalidStatus
		}
	}

	// Validate templates
	for _, template := range data.Templates {
		if template == nil {
			continue // Skip nil entries
		}
		if template.Id == "" {
			return ErrInvalidTemplate
		}
		if template.TxType == pb.TransactionType_TX_TYPE_UNSPECIFIED {
			return ErrInvalidTransactionType
		}
	}

	return nil
}
