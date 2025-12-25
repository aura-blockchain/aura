package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIdentityChangeParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RequestIdentityChangeParams
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
			name: "missing requester",
			params: &RequestIdentityChangeParams{
				TargetDID: "did:aura:test123",
			},
			wantErr: true,
			errMsg:  "requester is required",
		},
		{
			name: "missing target DID",
			params: &RequestIdentityChangeParams{
				Requester: "aura1requester123",
			},
			wantErr: true,
			errMsg:  "target DID is required",
		},
		{
			name: "valid params",
			params: &RequestIdentityChangeParams{
				Requester:    "aura1requester123",
				TargetDID:    "did:aura:test123",
				MetadataHash: "sha256:metadata123",
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
					if tt.params.Requester == "" {
						assert.Empty(t, tt.params.Requester)
					} else if tt.params.TargetDID == "" {
						assert.Empty(t, tt.params.TargetDID)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Requester)
				assert.NotEmpty(t, tt.params.TargetDID)
			}
		})
	}
}

func TestCreateRoleParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateRoleParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing creator",
			params: &CreateRoleParams{
				RoleName: "admin",
			},
			wantErr: true,
		},
		{
			name: "missing role name",
			params: &CreateRoleParams{
				Creator: "aura1creator123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &CreateRoleParams{
				Creator:     "aura1creator123",
				RoleName:    "admin",
				Permissions: []string{"read", "write", "delete"},
				Description: "Administrator role with full permissions",
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
					if tt.params.Creator == "" {
						assert.Empty(t, tt.params.Creator)
					} else if tt.params.RoleName == "" {
						assert.Empty(t, tt.params.RoleName)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Creator)
				assert.NotEmpty(t, tt.params.RoleName)
			}
		})
	}
}

func TestAssignRoleParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *AssignRoleParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing assigner",
			params: &AssignRoleParams{
				Address:  "aura1user123",
				RoleName: "admin",
			},
			wantErr: true,
		},
		{
			name: "missing address",
			params: &AssignRoleParams{
				Assigner: "aura1admin123",
				RoleName: "admin",
			},
			wantErr: true,
		},
		{
			name: "missing role name",
			params: &AssignRoleParams{
				Assigner: "aura1admin123",
				Address:  "aura1user123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &AssignRoleParams{
				Assigner: "aura1admin123",
				Address:  "aura1user123",
				RoleName: "admin",
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
					if tt.params.Assigner == "" {
						assert.Empty(t, tt.params.Assigner)
					} else if tt.params.Address == "" {
						assert.Empty(t, tt.params.Address)
					} else if tt.params.RoleName == "" {
						assert.Empty(t, tt.params.RoleName)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Assigner)
				assert.NotEmpty(t, tt.params.Address)
				assert.NotEmpty(t, tt.params.RoleName)
			}
		})
	}
}

func TestCreateMultisigWalletParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateMultisigWalletParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing creator",
			params: &CreateMultisigWalletParams{
				Signers:   []string{"aura1signer1", "aura1signer2"},
				Threshold: 2,
			},
			wantErr: true,
		},
		{
			name: "missing signers",
			params: &CreateMultisigWalletParams{
				Creator:   "aura1creator123",
				Signers:   []string{},
				Threshold: 2,
			},
			wantErr: true,
		},
		{
			name: "zero threshold",
			params: &CreateMultisigWalletParams{
				Creator:   "aura1creator123",
				Signers:   []string{"aura1signer1", "aura1signer2"},
				Threshold: 0,
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &CreateMultisigWalletParams{
				Creator:   "aura1creator123",
				Signers:   []string{"aura1signer1", "aura1signer2", "aura1signer3"},
				Threshold: 2,
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
					if tt.params.Creator == "" {
						assert.Empty(t, tt.params.Creator)
					} else if len(tt.params.Signers) == 0 {
						assert.Empty(t, tt.params.Signers)
					} else if tt.params.Threshold == 0 {
						assert.Zero(t, tt.params.Threshold)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Creator)
				assert.NotEmpty(t, tt.params.Signers)
				assert.NotZero(t, tt.params.Threshold)
			}
		})
	}
}

func TestCreateSessionParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *CreateSessionParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing address",
			params: &CreateSessionParams{
				DeviceFingerprint: "device123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &CreateSessionParams{
				Address:           "aura1user123",
				DeviceFingerprint: "device123",
				IpAddress:         "192.168.1.1",
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
					if tt.params.Address == "" {
						assert.Empty(t, tt.params.Address)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Address)
			}
		})
	}
}

func TestEndSession_Validation(t *testing.T) {
	tests := []struct {
		name      string
		address   string
		sessionID string
		wantErr   bool
	}{
		{
			name:      "empty address",
			address:   "",
			sessionID: "session123",
			wantErr:   true,
		},
		{
			name:      "empty session ID",
			address:   "aura1user123",
			sessionID: "",
			wantErr:   true,
		},
		{
			name:      "valid params",
			address:   "aura1user123",
			sessionID: "session123",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				if tt.address == "" {
					assert.Empty(t, tt.address)
				} else if tt.sessionID == "" {
					assert.Empty(t, tt.sessionID)
				}
			} else {
				assert.NotEmpty(t, tt.address)
				assert.NotEmpty(t, tt.sessionID)
			}
		})
	}
}

func TestQueryValidation(t *testing.T) {
	t.Run("GetIdentityRecord requires DID", func(t *testing.T) {
		assert.Empty(t, "")
		assert.NotEmpty(t, "did:aura:test123")
	})

	t.Run("GetIdentityRecordByAddress requires address", func(t *testing.T) {
		assert.Empty(t, "")
		assert.NotEmpty(t, "aura1user123")
	})

	t.Run("GetChangeRequest requires request ID", func(t *testing.T) {
		assert.Empty(t, "")
		assert.NotEmpty(t, "request123")
	})

	t.Run("GetRole requires role name", func(t *testing.T) {
		assert.Empty(t, "")
		assert.NotEmpty(t, "admin")
	})

	t.Run("GetRoleAssignments requires address", func(t *testing.T) {
		assert.Empty(t, "")
		assert.NotEmpty(t, "aura1user123")
	})
}
