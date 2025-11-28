package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	validAddress  = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
	validAddress2 = "cosmos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a"
)

func TestMsgCreateRole_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateRole
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCreateRole{
				Creator:     validAddress,
				Name:        "admin",
				Permissions: []string{"read", "write"},
				Description: "Admin role",
			},
			wantErr: false,
		},
		{
			name: "invalid creator address",
			msg: &MsgCreateRole{
				Creator:     "invalid",
				Name:        "admin",
				Permissions: []string{"read"},
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "empty name",
			msg: &MsgCreateRole{
				Creator:     validAddress,
				Name:        "",
				Permissions: []string{"read"},
			},
			wantErr: true,
			errMsg:  "name",
		},
		{
			name: "empty permissions",
			msg: &MsgCreateRole{
				Creator:     validAddress,
				Name:        "admin",
				Permissions: []string{},
			},
			wantErr: true,
			errMsg:  "permissions",
		},
		{
			name: "too many permissions",
			msg: &MsgCreateRole{
				Creator:     validAddress,
				Name:        "admin",
				Permissions: make([]string, 51),
			},
			wantErr: true,
			errMsg:  "cannot exceed 50 permissions",
		},
		{
			name: "valid with empty description",
			msg: &MsgCreateRole{
				Creator:     validAddress,
				Name:        "admin",
				Permissions: []string{"read", "write"},
				Description: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill permissions slice for the "too many" test
			if len(tt.msg.Permissions) == 51 {
				for i := range tt.msg.Permissions {
					tt.msg.Permissions[i] = "permission"
				}
			}
			err := tt.msg.ValidateBasic()
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

func TestMsgAssignRole_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgAssignRole
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgAssignRole{
				Assigner:          validAddress,
				Address:           validAddress2,
				RoleName:          "admin",
				ExpiresInSeconds:  3600,
			},
			wantErr: false,
		},
		{
			name: "invalid assigner",
			msg: &MsgAssignRole{
				Assigner:          "invalid",
				Address:           validAddress2,
				RoleName:          "admin",
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "assigner",
		},
		{
			name: "invalid target address",
			msg: &MsgAssignRole{
				Assigner:          validAddress,
				Address:           "invalid",
				RoleName:          "admin",
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "address",
		},
		{
			name: "empty role name",
			msg: &MsgAssignRole{
				Assigner:          validAddress,
				Address:           validAddress2,
				RoleName:          "",
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "role_name",
		},
		{
			name: "negative expiry",
			msg: &MsgAssignRole{
				Assigner:          validAddress,
				Address:           validAddress2,
				RoleName:          "admin",
				ExpiresInSeconds:  -1,
			},
			wantErr: true,
			errMsg:  "expires_in_seconds",
		},
		{
			name: "zero expiry (no expiry)",
			msg: &MsgAssignRole{
				Assigner:          validAddress,
				Address:           validAddress2,
				RoleName:          "admin",
				ExpiresInSeconds:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgRevokeRole_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRevokeRole
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgRevokeRole{
				Revoker:  validAddress,
				Address:  validAddress2,
				RoleName: "admin",
			},
			wantErr: false,
		},
		{
			name: "invalid revoker",
			msg: &MsgRevokeRole{
				Revoker:  "invalid",
				Address:  validAddress2,
				RoleName: "admin",
			},
			wantErr: true,
			errMsg:  "revoker",
		},
		{
			name: "empty role name",
			msg: &MsgRevokeRole{
				Revoker:  validAddress,
				Address:  validAddress2,
				RoleName: "",
			},
			wantErr: true,
			errMsg:  "role_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgCreateMultisigWallet_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateMultisigWallet
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid 3-of-5 wallet",
			msg: &MsgCreateMultisigWallet{
				Creator:    validAddress,
				Signers:    []string{validAddress, validAddress2, validAddress, validAddress2, validAddress},
				Threshold:  3,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: false,
		},
		{
			name: "invalid creator",
			msg: &MsgCreateMultisigWallet{
				Creator:    "invalid",
				Signers:    []string{validAddress, validAddress2},
				Threshold:  2,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: true,
			errMsg:  "creator",
		},
		{
			name: "no signers",
			msg: &MsgCreateMultisigWallet{
				Creator:    validAddress,
				Signers:    []string{},
				Threshold:  1,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: true,
			errMsg:  "must have at least",
		},
		{
			name: "threshold exceeds signers",
			msg: &MsgCreateMultisigWallet{
				Creator:    validAddress,
				Signers:    []string{validAddress, validAddress2},
				Threshold:  3,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: true,
			errMsg:  "threshold cannot exceed",
		},
		{
			name: "threshold zero",
			msg: &MsgCreateMultisigWallet{
				Creator:    validAddress,
				Signers:    []string{validAddress, validAddress2},
				Threshold:  0,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: true,
			errMsg:  "threshold",
		},
		{
			name: "invalid signer address",
			msg: &MsgCreateMultisigWallet{
				Creator:    validAddress,
				Signers:    []string{validAddress, "invalid"},
				Threshold:  2,
				WalletType: WalletType_WALLET_TYPE_CUSTOM,
			},
			wantErr: true,
			errMsg:  "signers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgCreateMultisigProposal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateMultisigProposal
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgCreateMultisigProposal{
				Proposer:          validAddress,
				WalletId:          "wallet-123",
				Title:             "Test Proposal",
				Description:       "This is a test proposal",
				Payload:           []byte("test payload"),
				ExpiresInSeconds:  86400,
			},
			wantErr: false,
		},
		{
			name: "invalid proposer",
			msg: &MsgCreateMultisigProposal{
				Proposer:          "invalid",
				WalletId:          "wallet-123",
				Title:             "Test Proposal",
				Description:       "This is a test proposal",
				Payload:           []byte("test payload"),
				ExpiresInSeconds:  86400,
			},
			wantErr: true,
			errMsg:  "proposer",
		},
		{
			name: "empty wallet ID",
			msg: &MsgCreateMultisigProposal{
				Proposer:          validAddress,
				WalletId:          "",
				Title:             "Test Proposal",
				Description:       "This is a test proposal",
				Payload:           []byte("test payload"),
				ExpiresInSeconds:  86400,
			},
			wantErr: true,
			errMsg:  "wallet_id",
		},
		{
			name: "empty title",
			msg: &MsgCreateMultisigProposal{
				Proposer:          validAddress,
				WalletId:          "wallet-123",
				Title:             "",
				Description:       "This is a test proposal",
				Payload:           []byte("test payload"),
				ExpiresInSeconds:  86400,
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "empty payload",
			msg: &MsgCreateMultisigProposal{
				Proposer:          validAddress,
				WalletId:          "wallet-123",
				Title:             "Test Proposal",
				Description:       "This is a test proposal",
				Payload:           []byte{},
				ExpiresInSeconds:  86400,
			},
			wantErr: true,
			errMsg:  "payload",
		},
		{
			name: "zero expiry",
			msg: &MsgCreateMultisigProposal{
				Proposer:          validAddress,
				WalletId:          "wallet-123",
				Title:             "Test Proposal",
				Description:       "This is a test proposal",
				Payload:           []byte("test payload"),
				ExpiresInSeconds:  0,
			},
			wantErr: true,
			errMsg:  "expires_in_seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgProposeTimeLockedAction_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgProposeTimeLockedAction
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgProposeTimeLockedAction{
				Proposer:     validAddress,
				ActionType:   "parameter_change",
				Payload:      []byte("test payload"),
				DelaySeconds: 86400, // 1 day
			},
			wantErr: false,
		},
		{
			name: "invalid proposer",
			msg: &MsgProposeTimeLockedAction{
				Proposer:     "invalid",
				ActionType:   "parameter_change",
				Payload:      []byte("test payload"),
				DelaySeconds: 86400,
			},
			wantErr: true,
			errMsg:  "proposer",
		},
		{
			name: "empty action type",
			msg: &MsgProposeTimeLockedAction{
				Proposer:     validAddress,
				ActionType:   "",
				Payload:      []byte("test payload"),
				DelaySeconds: 86400,
			},
			wantErr: true,
			errMsg:  "action_type",
		},
		{
			name: "delay too short",
			msg: &MsgProposeTimeLockedAction{
				Proposer:     validAddress,
				ActionType:   "parameter_change",
				Payload:      []byte("test payload"),
				DelaySeconds: 60, // < 1 hour
			},
			wantErr: true,
			errMsg:  "delay_seconds",
		},
		{
			name: "delay too long",
			msg: &MsgProposeTimeLockedAction{
				Proposer:     validAddress,
				ActionType:   "parameter_change",
				Payload:      []byte("test payload"),
				DelaySeconds: 31 * 24 * 60 * 60, // > 30 days
			},
			wantErr: true,
			errMsg:  "delay_seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgActivateEmergencyAdmin_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgActivateEmergencyAdmin
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgActivateEmergencyAdmin{
				Activator:         validAddress,
				AdminAddress:      validAddress2,
				Privileges:        []string{"pause_transfers", "emergency_withdraw"},
				ExpiresInSeconds:  3600,
			},
			wantErr: false,
		},
		{
			name: "invalid activator",
			msg: &MsgActivateEmergencyAdmin{
				Activator:         "invalid",
				AdminAddress:      validAddress2,
				Privileges:        []string{"pause_transfers"},
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "activator",
		},
		{
			name: "no privileges",
			msg: &MsgActivateEmergencyAdmin{
				Activator:         validAddress,
				AdminAddress:      validAddress2,
				Privileges:        []string{},
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "privileges",
		},
		{
			name: "too many privileges",
			msg: &MsgActivateEmergencyAdmin{
				Activator:         validAddress,
				AdminAddress:      validAddress2,
				Privileges:        make([]string, 51),
				ExpiresInSeconds:  3600,
			},
			wantErr: true,
			errMsg:  "cannot exceed 50 privileges",
		},
		{
			name: "zero expiry",
			msg: &MsgActivateEmergencyAdmin{
				Activator:         validAddress,
				AdminAddress:      validAddress2,
				Privileges:        []string{"pause_transfers"},
				ExpiresInSeconds:  0,
			},
			wantErr: true,
			errMsg:  "expires_in_seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill privileges slice for the "too many" test
			if len(tt.msg.Privileges) == 51 {
				for i := range tt.msg.Privileges {
					tt.msg.Privileges[i] = "privilege"
				}
			}
			err := tt.msg.ValidateBasic()
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

func TestMsgInitiateValidatorKeyRotation_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgInitiateValidatorKeyRotation
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgInitiateValidatorKeyRotation{
				Initiator:          validAddress,
				ValidatorAddress:   validAddress2,
				NewConsensusPubkey: "cosmosvalconspub1zcjduepqxxxxxx",
			},
			wantErr: false,
		},
		{
			name: "invalid initiator",
			msg: &MsgInitiateValidatorKeyRotation{
				Initiator:          "invalid",
				ValidatorAddress:   validAddress2,
				NewConsensusPubkey: "cosmosvalconspub1zcjduepqxxxxxx",
			},
			wantErr: true,
			errMsg:  "initiator",
		},
		{
			name: "empty consensus pubkey",
			msg: &MsgInitiateValidatorKeyRotation{
				Initiator:          validAddress,
				ValidatorAddress:   validAddress2,
				NewConsensusPubkey: "",
			},
			wantErr: true,
			errMsg:  "new_consensus_pubkey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgCreateSession_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgCreateSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message with IP",
			msg: &MsgCreateSession{
				UserAddress: validAddress,
				IpAddress:   "192.168.1.1",
				Metadata:    map[string]string{"browser": "chrome"},
			},
			wantErr: false,
		},
		{
			name: "valid message without IP",
			msg: &MsgCreateSession{
				UserAddress: validAddress,
				IpAddress:   "",
				Metadata:    nil,
			},
			wantErr: false,
		},
		{
			name: "invalid user address",
			msg: &MsgCreateSession{
				UserAddress: "invalid",
				IpAddress:   "192.168.1.1",
			},
			wantErr: true,
			errMsg:  "user_address",
		},
		{
			name: "IP address too short",
			msg: &MsgCreateSession{
				UserAddress: validAddress,
				IpAddress:   "1.1.1",
			},
			wantErr: true,
			errMsg:  "ip_address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgRevokeSession_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgRevokeSession
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgRevokeSession{
				UserAddress: validAddress,
				SessionId:   "session-123",
			},
			wantErr: false,
		},
		{
			name: "invalid user address",
			msg: &MsgRevokeSession{
				UserAddress: "invalid",
				SessionId:   "session-123",
			},
			wantErr: true,
			errMsg:  "user_address",
		},
		{
			name: "empty session ID",
			msg: &MsgRevokeSession{
				UserAddress: validAddress,
				SessionId:   "",
			},
			wantErr: true,
			errMsg:  "session_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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

func TestMsgSignMultisigProposal_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *MsgSignMultisigProposal
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: &MsgSignMultisigProposal{
				Signer:     validAddress,
				ProposalId: "proposal-123",
			},
			wantErr: false,
		},
		{
			name: "invalid signer",
			msg: &MsgSignMultisigProposal{
				Signer:     "invalid",
				ProposalId: "proposal-123",
			},
			wantErr: true,
			errMsg:  "signer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
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
