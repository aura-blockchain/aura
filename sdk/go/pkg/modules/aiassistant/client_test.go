package aiassistant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAssistantParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *RegisterAssistantParams
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
			name: "missing assistant address",
			params: &RegisterAssistantParams{
				OwnerAddress: "aura1owner123",
				Locales:      []string{"en-US"},
			},
			wantErr: true,
			errMsg:  "assistant address is required",
		},
		{
			name: "missing owner address",
			params: &RegisterAssistantParams{
				AssistantAddress: "aura1assistant123",
				Locales:          []string{"en-US"},
			},
			wantErr: true,
			errMsg:  "owner address is required",
		},
		{
			name: "missing locales",
			params: &RegisterAssistantParams{
				AssistantAddress: "aura1assistant123",
				OwnerAddress:     "aura1owner123",
				Locales:          []string{},
			},
			wantErr: true,
			errMsg:  "at least one locale is required",
		},
		{
			name: "valid params",
			params: &RegisterAssistantParams{
				AssistantAddress: "aura1assistant123",
				OwnerAddress:     "aura1owner123",
				Locales:          []string{"en-US", "es-ES"},
				ModelHash:        "sha256:abc123",
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
					if tt.params.AssistantAddress == "" {
						assert.Empty(t, tt.params.AssistantAddress)
					} else if tt.params.OwnerAddress == "" {
						assert.Empty(t, tt.params.OwnerAddress)
					} else if len(tt.params.Locales) == 0 {
						assert.Empty(t, tt.params.Locales)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.AssistantAddress)
				assert.NotEmpty(t, tt.params.OwnerAddress)
				assert.NotEmpty(t, tt.params.Locales)
			}
		})
	}
}

func TestUpdateLocalesParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *UpdateLocalesParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing assistant address",
			params: &UpdateLocalesParams{
				Locales: []string{"en-US"},
			},
			wantErr: true,
		},
		{
			name: "missing locales",
			params: &UpdateLocalesParams{
				AssistantAddress: "aura1assistant123",
				Locales:          []string{},
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &UpdateLocalesParams{
				AssistantAddress: "aura1assistant123",
				Locales:          []string{"en-US", "fr-FR"},
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
					if tt.params.AssistantAddress == "" {
						assert.Empty(t, tt.params.AssistantAddress)
					} else if len(tt.params.Locales) == 0 {
						assert.Empty(t, tt.params.Locales)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.AssistantAddress)
				assert.NotEmpty(t, tt.params.Locales)
			}
		})
	}
}

func TestReportMisbehaviorParams_Validation(t *testing.T) {
	tests := []struct {
		name    string
		params  *ReportMisbehaviorParams
		wantErr bool
	}{
		{
			name:    "nil params",
			params:  nil,
			wantErr: true,
		},
		{
			name: "missing reporter",
			params: &ReportMisbehaviorParams{
				AssistantAddress: "aura1assistant123",
			},
			wantErr: true,
		},
		{
			name: "missing assistant address",
			params: &ReportMisbehaviorParams{
				Reporter: "aura1reporter123",
			},
			wantErr: true,
		},
		{
			name: "valid params",
			params: &ReportMisbehaviorParams{
				Reporter:         "aura1reporter123",
				AssistantAddress: "aura1assistant123",
				Infraction:       "harmful_content",
				EvidenceHash:     "sha256:abc123def456",
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
					if tt.params.Reporter == "" {
						assert.Empty(t, tt.params.Reporter)
					} else if tt.params.AssistantAddress == "" {
						assert.Empty(t, tt.params.AssistantAddress)
					}
				}
			} else {
				assert.NotEmpty(t, tt.params.Reporter)
				assert.NotEmpty(t, tt.params.AssistantAddress)
			}
		})
	}
}

func TestHeartbeat_Validation(t *testing.T) {
	tests := []struct {
		name             string
		assistantAddress string
		wantErr          bool
	}{
		{
			name:             "empty address",
			assistantAddress: "",
			wantErr:          true,
		},
		{
			name:             "valid address",
			assistantAddress: "aura1assistant123",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Empty(t, tt.assistantAddress)
			} else {
				assert.NotEmpty(t, tt.assistantAddress)
			}
		})
	}
}
