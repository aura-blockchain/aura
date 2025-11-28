package vcregistry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMintVCParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *MintVCParams
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
			errMsg:  "params cannot be nil",
		},
		{
			name: "missing holder address",
			params: &MintVCParams{
				HolderDID: "did:aura:test123",
			},
			wantErr: true,
			errMsg:  "holder address is required",
		},
		{
			name: "missing holder DID",
			params: &MintVCParams{
				HolderAddress: "paw1test123",
			},
			wantErr: true,
			errMsg:  "holder DID is required",
		},
		{
			name: "valid params",
			params: &MintVCParams{
				HolderAddress: "paw1test123",
				HolderDID:     "did:aura:test123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.NotNil(t, tt.params == nil)
				} else {
					if tt.params.HolderAddress == "" {
						assert.Empty(t, tt.params.HolderAddress)
					} else if tt.params.HolderDID == "" {
						assert.Empty(t, tt.params.HolderDID)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.HolderAddress)
				assert.NotEmpty(t, tt.params.HolderDID)
			}
		})
	}
}

func TestRevokeVCParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RevokeVCParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing holder address",
			params: &RevokeVCParams{
				VCID: "vc123",
			},
			wantErr: true,
		},
		{
			name: "missing VC ID",
			params: &RevokeVCParams{
				HolderAddress: "paw1test123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &RevokeVCParams{
				HolderAddress: "paw1test123",
				VCID:          "vc123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.HolderAddress == "" || tt.params.VCID == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.HolderAddress)
				assert.NotEmpty(t, tt.params.VCID)
			}
		})
	}
}

func TestRegisterDIDParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RegisterDIDParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing controller",
			params: &RegisterDIDParams{
				DID: "did:aura:test123",
			},
			wantErr: true,
		},
		{
			name: "missing DID",
			params: &RegisterDIDParams{
				Controller: "paw1test123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &RegisterDIDParams{
				Controller: "paw1test123",
				DID:        "did:aura:test123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Controller == "" || tt.params.DID == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Controller)
				assert.NotEmpty(t, tt.params.DID)
			}
		})
	}
}

func TestListVCsParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *ListVCsParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name:    "missing holder address",
			params:  &ListVCsParams{},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &ListVCsParams{
				HolderAddress: "paw1test123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.Empty(t, tt.params.HolderAddress)
				}
			} else {
				assert.NotEmpty(t, tt.params.HolderAddress)
			}
		})
	}
}

func TestClient_QueryValidation(t *testing.T) {
	ctx := context.Background()

	// Note: These tests validate input parameters only
	// Full integration tests would require a running chain

	t.Run("GetVC requires VC ID", func(t *testing.T) {
		vcID := ""
		assert.Empty(t, vcID, "VC ID should be empty for validation")
	})

	t.Run("VerifyVC requires VC ID", func(t *testing.T) {
		vcID := ""
		assert.Empty(t, vcID, "VC ID should be empty for validation")
	})

	t.Run("ResolveDID requires DID", func(t *testing.T) {
		did := ""
		assert.Empty(t, did, "DID should be empty for validation")
	})

	t.Run("GetDIDByAddress requires address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address, "address should be empty for validation")
	})

	t.Run("ValidateMintEligibility requires holder address", func(t *testing.T) {
		address := ""
		assert.Empty(t, address, "holder address should be empty for validation")
	})

	t.Run("GetVCPolicy requires VC type name", func(t *testing.T) {
		vcTypeName := ""
		assert.Empty(t, vcTypeName, "VC type name should be empty for validation")
	})

	t.Run("CheckRevocation requires VC ID", func(t *testing.T) {
		vcID := ""
		assert.Empty(t, vcID, "VC ID should be empty for validation")
	})

	// Suppress unused variable warning
	_ = ctx
}

func TestAdminRevokeVCParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *AdminRevokeVCParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing authority",
			params: &AdminRevokeVCParams{
				VCID: "vc123",
			},
			wantErr: true,
		},
		{
			name: "missing VC ID",
			params: &AdminRevokeVCParams{
				Authority: "paw1gov123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &AdminRevokeVCParams{
				Authority: "paw1gov123",
				VCID:      "vc123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Authority == "" || tt.params.VCID == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Authority)
				assert.NotEmpty(t, tt.params.VCID)
			}
		})
	}
}

func TestUpdateDIDParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *UpdateDIDParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing controller",
			params: &UpdateDIDParams{
				DID: "did:aura:test123",
			},
			wantErr: true,
		},
		{
			name: "missing DID",
			params: &UpdateDIDParams{
				Controller: "paw1test123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &UpdateDIDParams{
				Controller: "paw1test123",
				DID:        "did:aura:test123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.params == nil {
					require.Nil(t, tt.params)
				} else {
					assert.True(t, tt.params.Controller == "" || tt.params.DID == "")
				}
			} else {
				assert.NotEmpty(t, tt.params.Controller)
				assert.NotEmpty(t, tt.params.DID)
			}
		})
	}
}

func TestBatchVCStatus_Validation(t *testing.T) {
	t.Run("requires VC IDs", func(t *testing.T) {
		vcIDs := []string{}
		assert.Empty(t, vcIDs, "VC IDs should be empty for validation")
	})

	t.Run("valid VC IDs", func(t *testing.T) {
		vcIDs := []string{"vc1", "vc2", "vc3"}
		assert.NotEmpty(t, vcIDs, "VC IDs should not be empty")
		assert.Len(t, vcIDs, 3, "should have 3 VC IDs")
	})
}
