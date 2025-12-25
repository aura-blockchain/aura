// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	sdkmath "cosmossdk.io/math"
	pb "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// GenesisState defines the privacy module's genesis state
type GenesisState struct {
	Params Params `json:"params"`
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// DefaultGenesisState returns the default genesis state in proto format
func DefaultGenesisState() *pb.GenesisState {
	return &pb.GenesisState{
		Params: pb.Params{
			EnableZkProofs:                 true,
			EnableStealthAddresses:         true,
			EnableRingSignatures:           true,
			EnableConfidentialTransactions: true,
			EnableNetworkPrivacy:           true,
			EnableMixing:                   true,
			MinRingSize:                    3,
			MaxRingSize:                    7,
			MinMixingParticipants:          2,
			MixingFee:                      sdkmath.NewInt(100),
			ZkProofVerificationCost:        50,
		},
		MixingPools:        make([]*pb.MixingPool, 0),
		RegisteredViewKeys: make([]*pb.ViewKey, 0),
	}
}

// Validate validates the genesis state
func (gs GenesisState) Validate() error {
	return ValidateParams(gs.Params)
}
