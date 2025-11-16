package types

import "fmt"

// Params defines the parameters for the dataregistry module
type Params struct {
	MaxDataItemsPerUser uint64   `json:"max_data_items_per_user"`
	MaxStorageBytes     uint64   `json:"max_storage_bytes"`
	StorageFee          string   `json:"storage_fee"`
	VerificationReward  uint64   `json:"verification_reward"`
	AuthorizedVerifiers []string `json:"authorized_verifiers"`
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		MaxDataItemsPerUser: 1000,
		MaxStorageBytes:     104857600, // 100MB
		StorageFee:          "100000uaura",
		VerificationReward:  1000,
		AuthorizedVerifiers: []string{},
	}
}

// Validate performs validation on the Params
func (p Params) Validate() error {
	if p.MaxDataItemsPerUser == 0 {
		return fmt.Errorf("max data items per user must be positive")
	}

	if p.MaxStorageBytes == 0 {
		return fmt.Errorf("max storage bytes must be positive")
	}

	if p.StorageFee == "" {
		return fmt.Errorf("storage fee cannot be empty")
	}

	return nil
}
