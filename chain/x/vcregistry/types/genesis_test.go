// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	pb "github.com/aequitas/aura/proto/aura/vcregistry/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesisState(t *testing.T) {
	gs := DefaultGenesisState()

	// Verify genesis state is valid
	err := ValidateGenesisState(gs)
	require.NoError(t, err, "default genesis state should be valid")

	// Verify non-nil fields
	require.NotNil(t, gs.Params, "params should not be nil")
	require.NotNil(t, gs.VcRecords, "vc_records should not be nil")
	require.NotNil(t, gs.RevocationRecords, "revocation_records should not be nil")
	require.NotNil(t, gs.RevocationList, "revocation_list should not be nil")
	require.NotNil(t, gs.DidDocuments, "did_documents should not be nil")
	require.NotNil(t, gs.VcPolicies, "vc_policies should not be nil")
}

func TestGenesisStateValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *pb.GenesisState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid default genesis",
			setup: func() *pb.GenesisState {
				return DefaultGenesisState()
			},
			wantErr: false,
		},
		{
			name: "nil genesis state",
			setup: func() *pb.GenesisState {
				return nil
			},
			wantErr: true,
			errMsg:  "genesis state cannot be nil",
		},
		{
			name: "invalid params - negative max VCs",
			setup: func() *pb.GenesisState {
				gs := DefaultGenesisState()
				gs.Params.MaxVcsPerUser = 0 // Invalid: must be positive
				return gs
			},
			wantErr: true,
			errMsg:  "invalid params",
		},
		{
			name: "nil revocation list",
			setup: func() *pb.GenesisState {
				gs := DefaultGenesisState()
				gs.RevocationList = nil
				return gs
			},
			wantErr: true,
			errMsg:  "revocation_list cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gs := tt.setup()
			err := ValidateGenesisState(gs)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
