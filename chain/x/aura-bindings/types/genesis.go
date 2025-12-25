// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"fmt"
)

// GenesisState defines the aura-bindings module's genesis state
type GenesisState struct {
	// QueryStats tracks query usage statistics
	QueryStats map[string]uint64 `json:"query_stats"`

	// MessageStats tracks message usage statistics
	MessageStats map[string]uint64 `json:"message_stats"`
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		QueryStats:   make(map[string]uint64),
		MessageStats: make(map[string]uint64),
	}
}

// Validate performs basic validation of genesis data
func (gs GenesisState) Validate() error {
	if gs.QueryStats == nil {
		return fmt.Errorf("query stats cannot be nil")
	}

	if gs.MessageStats == nil {
		return fmt.Errorf("message stats cannot be nil")
	}

	// Validate that stats are non-negative (always true for uint64, but explicit check)
	for queryType, count := range gs.QueryStats {
		if queryType == "" {
			return fmt.Errorf("query type cannot be empty")
		}
		_ = count // count >= 0 always true for uint64
	}

	for msgType, count := range gs.MessageStats {
		if msgType == "" {
			return fmt.Errorf("message type cannot be empty")
		}
		_ = count // count >= 0 always true for uint64
	}

	return nil
}
