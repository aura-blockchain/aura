package types

import (
	"testing"
	"time"

	authproto "github.com/aequitas/aura/proto/aura/auth/v1beta1"
	"github.com/stretchr/testify/require"
)

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name string
		role *authproto.Role
		err  string
	}{
		{
			name: "valid role",
			role: &authproto.Role{
				Name:        "admin",
				Permissions: []string{PermissionAdmin, PermissionCreateRole},
			},
			err: "",
		},
		{
			name: "empty name",
			role: &authproto.Role{
				Name:        "",
				Permissions: []string{PermissionAdmin},
			},
			err: "role name cannot be empty",
		},
		{
			name: "no permissions",
			role: &authproto.Role{
				Name:        "admin",
				Permissions: []string{},
			},
			err: "role must have at least one permission",
		},
		{
			name: "nil permissions",
			role: &authproto.Role{
				Name:        "admin",
				Permissions: nil,
			},
			err: "role must have at least one permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateRoleAssignment(t *testing.T) {
	tests := []struct {
		name       string
		assignment *authproto.RoleAssignment
		err        string
	}{
		{
			name: "valid assignment",
			assignment: &authproto.RoleAssignment{
				Address:    "aura1abc123",
				RoleName:   RoleAdmin,
				AssignedBy: "aura1xyz789",
			},
			err: "",
		},
		{
			name: "empty address",
			assignment: &authproto.RoleAssignment{
				Address:    "",
				RoleName:   RoleAdmin,
				AssignedBy: "aura1xyz789",
			},
			err: "address cannot be empty",
		},
		{
			name: "empty role name",
			assignment: &authproto.RoleAssignment{
				Address:    "aura1abc123",
				RoleName:   "",
				AssignedBy: "aura1xyz789",
			},
			err: "role name cannot be empty",
		},
		{
			name: "empty assigned by",
			assignment: &authproto.RoleAssignment{
				Address:    "aura1abc123",
				RoleName:   RoleAdmin,
				AssignedBy: "",
			},
			err: "assigned_by cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoleAssignment(tt.assignment)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMultisigWallet(t *testing.T) {
	tests := []struct {
		name   string
		wallet *authproto.MultisigWallet
		err    string
	}{
		{
			name: "valid custom wallet",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"signer1", "signer2", "signer3"},
				Threshold:  2,
				WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
			},
			err: "",
		},
		{
			name: "valid 3-of-5 wallet",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-2",
				Signers:    []string{"s1", "s2", "s3", "s4", "s5"},
				Threshold:  3,
				WalletType: authproto.WalletType_WALLET_TYPE_3_OF_5,
			},
			err: "",
		},
		{
			name: "valid 5-of-7 wallet",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-3",
				Signers:    []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"},
				Threshold:  5,
				WalletType: authproto.WalletType_WALLET_TYPE_5_OF_7,
			},
			err: "",
		},
		{
			name: "empty ID",
			wallet: &authproto.MultisigWallet{
				Id:         "",
				Signers:    []string{"signer1", "signer2"},
				Threshold:  2,
				WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
			},
			err: "wallet ID cannot be empty",
		},
		{
			name: "no signers",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{},
				Threshold:  1,
				WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
			},
			err: "wallet must have at least one signer",
		},
		{
			name: "zero threshold",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"signer1", "signer2"},
				Threshold:  0,
				WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
			},
			err: "threshold must be greater than 0",
		},
		{
			name: "threshold exceeds signers",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"signer1", "signer2"},
				Threshold:  3,
				WalletType: authproto.WalletType_WALLET_TYPE_CUSTOM,
			},
			err: "threshold cannot be greater than number of signers",
		},
		{
			name: "invalid 3-of-5 config - wrong signers",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"s1", "s2", "s3"},
				Threshold:  3,
				WalletType: authproto.WalletType_WALLET_TYPE_3_OF_5,
			},
			err: "3-of-5 wallet must have 5 signers and threshold of 3",
		},
		{
			name: "invalid 3-of-5 config - wrong threshold",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"s1", "s2", "s3", "s4", "s5"},
				Threshold:  2,
				WalletType: authproto.WalletType_WALLET_TYPE_3_OF_5,
			},
			err: "3-of-5 wallet must have 5 signers and threshold of 3",
		},
		{
			name: "invalid 5-of-7 config",
			wallet: &authproto.MultisigWallet{
				Id:         "wallet-1",
				Signers:    []string{"s1", "s2", "s3", "s4", "s5"},
				Threshold:  5,
				WalletType: authproto.WalletType_WALLET_TYPE_5_OF_7,
			},
			err: "5-of-7 wallet must have 7 signers and threshold of 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMultisigWallet(tt.wallet)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMultisigProposal(t *testing.T) {
	tests := []struct {
		name     string
		proposal *authproto.MultisigProposal
		err      string
	}{
		{
			name: "valid proposal",
			proposal: &authproto.MultisigProposal{
				Id:       "prop-1",
				WalletId: "wallet-1",
				Title:    "Test Proposal",
				Payload:  []byte("payload"),
			},
			err: "",
		},
		{
			name: "empty ID",
			proposal: &authproto.MultisigProposal{
				Id:       "",
				WalletId: "wallet-1",
				Title:    "Test",
				Payload:  []byte("payload"),
			},
			err: "proposal ID cannot be empty",
		},
		{
			name: "empty wallet ID",
			proposal: &authproto.MultisigProposal{
				Id:       "prop-1",
				WalletId: "",
				Title:    "Test",
				Payload:  []byte("payload"),
			},
			err: "wallet ID cannot be empty",
		},
		{
			name: "empty title",
			proposal: &authproto.MultisigProposal{
				Id:       "prop-1",
				WalletId: "wallet-1",
				Title:    "",
				Payload:  []byte("payload"),
			},
			err: "title cannot be empty",
		},
		{
			name: "empty payload",
			proposal: &authproto.MultisigProposal{
				Id:       "prop-1",
				WalletId: "wallet-1",
				Title:    "Test",
				Payload:  []byte{},
			},
			err: "payload cannot be empty",
		},
		{
			name: "nil payload",
			proposal: &authproto.MultisigProposal{
				Id:       "prop-1",
				WalletId: "wallet-1",
				Title:    "Test",
				Payload:  nil,
			},
			err: "payload cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMultisigProposal(tt.proposal)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsProposalApproved(t *testing.T) {
	tests := []struct {
		name     string
		proposal *authproto.MultisigProposal
		wallet   *authproto.MultisigWallet
		expected bool
	}{
		{
			name: "approved",
			proposal: &authproto.MultisigProposal{
				Signatures: []string{"sig1", "sig2", "sig3"},
			},
			wallet: &authproto.MultisigWallet{
				Threshold: 3,
			},
			expected: true,
		},
		{
			name: "not approved",
			proposal: &authproto.MultisigProposal{
				Signatures: []string{"sig1", "sig2"},
			},
			wallet: &authproto.MultisigWallet{
				Threshold: 3,
			},
			expected: false,
		},
		{
			name: "over threshold",
			proposal: &authproto.MultisigProposal{
				Signatures: []string{"sig1", "sig2", "sig3", "sig4"},
			},
			wallet: &authproto.MultisigWallet{
				Threshold: 3,
			},
			expected: true,
		},
		{
			name: "no signatures",
			proposal: &authproto.MultisigProposal{
				Signatures: []string{},
			},
			wallet: &authproto.MultisigWallet{
				Threshold: 1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProposalApproved(tt.proposal, tt.wallet)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestIsProposalExpired(t *testing.T) {
	pastTime := time.Now().Add(-time.Hour)
	futureTime := time.Now().Add(time.Hour)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		proposal *authproto.MultisigProposal
		expected bool
	}{
		{
			name: "expired",
			proposal: &authproto.MultisigProposal{
				ExpiresAt: pastTime,
			},
			expected: true,
		},
		{
			name: "not expired",
			proposal: &authproto.MultisigProposal{
				ExpiresAt: futureTime,
			},
			expected: false,
		},
		{
			name: "no expiration",
			proposal: &authproto.MultisigProposal{
				ExpiresAt: zeroTime,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProposalExpired(tt.proposal)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateTimeLockedAction(t *testing.T) {
	tests := []struct {
		name   string
		action *authproto.TimeLockedAction
		err    string
	}{
		{
			name: "valid action",
			action: &authproto.TimeLockedAction{
				Id:           "action-1",
				ActionType:   "transfer",
				Payload:      []byte("payload"),
				Proposer:     "aura1abc123",
				DelaySeconds: 3600,
			},
			err: "",
		},
		{
			name: "empty ID",
			action: &authproto.TimeLockedAction{
				Id:           "",
				ActionType:   "transfer",
				Payload:      []byte("payload"),
				Proposer:     "aura1abc123",
				DelaySeconds: 3600,
			},
			err: "action ID cannot be empty",
		},
		{
			name: "empty action type",
			action: &authproto.TimeLockedAction{
				Id:           "action-1",
				ActionType:   "",
				Payload:      []byte("payload"),
				Proposer:     "aura1abc123",
				DelaySeconds: 3600,
			},
			err: "action type cannot be empty",
		},
		{
			name: "empty payload",
			action: &authproto.TimeLockedAction{
				Id:           "action-1",
				ActionType:   "transfer",
				Payload:      []byte{},
				Proposer:     "aura1abc123",
				DelaySeconds: 3600,
			},
			err: "payload cannot be empty",
		},
		{
			name: "empty proposer",
			action: &authproto.TimeLockedAction{
				Id:           "action-1",
				ActionType:   "transfer",
				Payload:      []byte("payload"),
				Proposer:     "",
				DelaySeconds: 3600,
			},
			err: "proposer cannot be empty",
		},
		{
			name: "zero delay",
			action: &authproto.TimeLockedAction{
				Id:           "action-1",
				ActionType:   "transfer",
				Payload:      []byte("payload"),
				Proposer:     "aura1abc123",
				DelaySeconds: 0,
			},
			err: "delay must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeLockedAction(tt.action)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsActionReady(t *testing.T) {
	pastTime := time.Now().Add(-time.Hour)
	futureTime := time.Now().Add(time.Hour)
	nowTime := time.Now()
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		action   *authproto.TimeLockedAction
		expected bool
	}{
		{
			name: "ready - past time",
			action: &authproto.TimeLockedAction{
				ExecutableAt: pastTime,
			},
			expected: true,
		},
		{
			name: "ready - exact time",
			action: &authproto.TimeLockedAction{
				ExecutableAt: nowTime,
			},
			expected: true,
		},
		{
			name: "not ready - future time",
			action: &authproto.TimeLockedAction{
				ExecutableAt: futureTime,
			},
			expected: false,
		},
		{
			name: "not ready - no time set",
			action: &authproto.TimeLockedAction{
				ExecutableAt: zeroTime,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsActionReady(tt.action)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEmergencyAdmin(t *testing.T) {
	tests := []struct {
		name  string
		admin *authproto.EmergencyAdmin
		err   string
	}{
		{
			name: "valid admin",
			admin: &authproto.EmergencyAdmin{
				Address:     "aura1abc123",
				Privileges:  []string{"pause_chain", "upgrade"},
				ActivatedBy: "aura1xyz789",
			},
			err: "",
		},
		{
			name: "empty address",
			admin: &authproto.EmergencyAdmin{
				Address:     "",
				Privileges:  []string{"pause_chain"},
				ActivatedBy: "aura1xyz789",
			},
			err: "address cannot be empty",
		},
		{
			name: "no privileges",
			admin: &authproto.EmergencyAdmin{
				Address:     "aura1abc123",
				Privileges:  []string{},
				ActivatedBy: "aura1xyz789",
			},
			err: "emergency admin must have at least one privilege",
		},
		{
			name: "empty activated by",
			admin: &authproto.EmergencyAdmin{
				Address:     "aura1abc123",
				Privileges:  []string{"pause_chain"},
				ActivatedBy: "",
			},
			err: "activated_by cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmergencyAdmin(tt.admin)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsEmergencyAdminActive(t *testing.T) {
	pastTime := time.Now().Add(-time.Hour)
	futureTime := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		admin    *authproto.EmergencyAdmin
		expected bool
	}{
		{
			name: "active - no expiration",
			admin: &authproto.EmergencyAdmin{
				IsActive:  true,
				ExpiresAt: nil,
			},
			expected: true,
		},
		{
			name: "active - not expired",
			admin: &authproto.EmergencyAdmin{
				IsActive:  true,
				ExpiresAt: &futureTime,
			},
			expected: true,
		},
		{
			name: "inactive - expired",
			admin: &authproto.EmergencyAdmin{
				IsActive:  true,
				ExpiresAt: &pastTime,
			},
			expected: false,
		},
		{
			name: "inactive - marked inactive",
			admin: &authproto.EmergencyAdmin{
				IsActive:  false,
				ExpiresAt: &futureTime,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmergencyAdminActive(tt.admin)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateSession(t *testing.T) {
	tests := []struct {
		name    string
		session *authproto.Session
		err     string
	}{
		{
			name: "valid session",
			session: &authproto.Session{
				SessionId:   "session-1",
				UserAddress: "aura1abc123",
			},
			err: "",
		},
		{
			name: "empty session ID",
			session: &authproto.Session{
				SessionId:   "",
				UserAddress: "aura1abc123",
			},
			err: "session ID cannot be empty",
		},
		{
			name: "empty user address",
			session: &authproto.Session{
				SessionId:   "session-1",
				UserAddress: "",
			},
			err: "user address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSession(tt.session)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsSessionActive(t *testing.T) {
	pastTime := time.Now().Add(-time.Hour)
	futureTime := time.Now().Add(time.Hour)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		session  *authproto.Session
		expected bool
	}{
		{
			name: "active - not expired",
			session: &authproto.Session{
				IsActive:  true,
				ExpiresAt: futureTime,
			},
			expected: true,
		},
		{
			name: "inactive - expired",
			session: &authproto.Session{
				IsActive:  true,
				ExpiresAt: pastTime,
			},
			expected: false,
		},
		{
			name: "inactive - marked inactive",
			session: &authproto.Session{
				IsActive:  false,
				ExpiresAt: futureTime,
			},
			expected: false,
		},
		{
			name: "inactive - no expiration",
			session: &authproto.Session{
				IsActive:  true,
				ExpiresAt: zeroTime,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSessionActive(tt.session)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRateLimitConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *authproto.RateLimitConfig
		err    string
	}{
		{
			name: "valid config",
			config: &authproto.RateLimitConfig{
				UserAddress: "aura1abc123",
			},
			err: "",
		},
		{
			name: "empty user address",
			config: &authproto.RateLimitConfig{
				UserAddress: "",
			},
			err: "user address cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRateLimitConfig(tt.config)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	now := time.Now()
	minuteAgo := now.Add(-30 * time.Second)
	hourAgo := now.Add(-30 * time.Minute)
	dayAgo := now.Add(-12 * time.Hour)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		config   *authproto.RateLimitConfig
		expected bool
	}{
		{
			name: "not limited - under all limits",
			config: &authproto.RateLimitConfig{
				WindowStart:        minuteAgo,
				RequestsPerMinute:  60,
				CurrentMinuteCount: 30,
				RequestsPerHour:    3600,
				CurrentHourCount:   100,
				RequestsPerDay:     86400,
				CurrentDayCount:    1000,
			},
			expected: false,
		},
		{
			name: "limited - minute exceeded",
			config: &authproto.RateLimitConfig{
				WindowStart:        minuteAgo,
				RequestsPerMinute:  60,
				CurrentMinuteCount: 60,
				RequestsPerHour:    3600,
				CurrentHourCount:   100,
				RequestsPerDay:     86400,
				CurrentDayCount:    1000,
			},
			expected: true,
		},
		{
			name: "limited - hour exceeded",
			config: &authproto.RateLimitConfig{
				WindowStart:        hourAgo,
				RequestsPerMinute:  60,
				CurrentMinuteCount: 30,
				RequestsPerHour:    3600,
				CurrentHourCount:   3600,
				RequestsPerDay:     86400,
				CurrentDayCount:    5000,
			},
			expected: true,
		},
		{
			name: "limited - day exceeded",
			config: &authproto.RateLimitConfig{
				WindowStart:        dayAgo,
				RequestsPerMinute:  60,
				CurrentMinuteCount: 30,
				RequestsPerHour:    3600,
				CurrentHourCount:   100,
				RequestsPerDay:     86400,
				CurrentDayCount:    86400,
			},
			expected: true,
		},
		{
			name: "not limited - no window start",
			config: &authproto.RateLimitConfig{
				WindowStart:        zeroTime,
				RequestsPerMinute:  60,
				CurrentMinuteCount: 100,
				RequestsPerHour:    3600,
				CurrentHourCount:   5000,
				RequestsPerDay:     86400,
				CurrentDayCount:    100000,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRateLimited(tt.config)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.NotNil(t, params)
	require.Equal(t, uint64(3600), params.SessionTimeoutSeconds)
	require.Equal(t, uint64(86400), params.DefaultTimelockDelaySeconds)
	require.Equal(t, uint64(60), params.DefaultRequestsPerMinute)
	require.Equal(t, uint64(3600), params.DefaultRequestsPerHour)
	require.Equal(t, uint64(86400), params.DefaultRequestsPerDay)
	require.Equal(t, uint64(604800), params.MultisigProposalExpirySeconds)
	require.True(t, params.AuditLoggingEnabled)
}

func TestGenerateID(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		components []string
	}{
		{
			name:       "simple ID",
			prefix:     "role",
			components: []string{"admin"},
		},
		{
			name:       "multi-component ID",
			prefix:     "wallet",
			components: []string{"user1", "user2", "user3"},
		},
		{
			name:       "no components",
			prefix:     "session",
			components: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GenerateID(tt.prefix, tt.components...)
			require.NotEmpty(t, id)
			require.Contains(t, id, tt.prefix+"-")
			require.Greater(t, len(id), len(tt.prefix)+1)
		})
	}

	// Test uniqueness
	id1 := GenerateID("test", "component")
	time.Sleep(1 * time.Millisecond)
	id2 := GenerateID("test", "component")
	require.NotEqual(t, id1, id2, "Generated IDs should be unique")
}
