package mocks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockVCRegistryKeeper implements a mock VCRegistryKeeper for testing
type MockVCRegistryKeeper struct {
	Credentials        map[string]interface{} // Store VCs by ID
	ConfidenceScores   map[string]int64       // Store confidence scores by address
	IRScores           map[string]uint64      // Store IR scores by address
	VerificationStatus map[string]bool        // Store verification status by address
}

// NewMockVCRegistryKeeper creates a new mock VC registry keeper
func NewMockVCRegistryKeeper() *MockVCRegistryKeeper {
	return &MockVCRegistryKeeper{
		Credentials:        make(map[string]interface{}),
		ConfidenceScores:   make(map[string]int64),
		IRScores:           make(map[string]uint64),
		VerificationStatus: make(map[string]bool),
	}
}

// GetCredential returns a credential by ID
func (m *MockVCRegistryKeeper) GetCredential(ctx sdk.Context, id string) (interface{}, bool) {
	cred, ok := m.Credentials[id]
	return cred, ok
}

// SetCredential sets a credential
func (m *MockVCRegistryKeeper) SetCredential(ctx sdk.Context, id string, credential interface{}) {
	m.Credentials[id] = credential
}

// GetConfidenceScore returns the confidence score for an address
func (m *MockVCRegistryKeeper) GetConfidenceScore(ctx sdk.Context, address sdk.AccAddress) int64 {
	if score, ok := m.ConfidenceScores[address.String()]; ok {
		return score
	}
	return 0
}

// SetConfidenceScore sets the confidence score for an address
func (m *MockVCRegistryKeeper) SetConfidenceScore(ctx sdk.Context, address sdk.AccAddress, score int64) {
	m.ConfidenceScores[address.String()] = score
}

// GetIRScore returns the IR (Inclusion Routine) score for an address
func (m *MockVCRegistryKeeper) GetIRScore(ctx sdk.Context, address string) uint64 {
	if score, ok := m.IRScores[address]; ok {
		return score
	}
	return 0
}

// SetIRScore sets the IR score for an address (test helper)
func (m *MockVCRegistryKeeper) SetIRScore(ctx sdk.Context, address string, score uint64) {
	m.IRScores[address] = score
}

// IsVerified returns whether an address is verified
func (m *MockVCRegistryKeeper) IsVerified(ctx sdk.Context, address string) bool {
	if verified, ok := m.VerificationStatus[address]; ok {
		return verified
	}
	return false
}

// SetVerified sets the verification status for an address
func (m *MockVCRegistryKeeper) SetVerified(ctx sdk.Context, address string, verified bool) {
	m.VerificationStatus[address] = verified
}

// HasCredential checks if a credential exists
func (m *MockVCRegistryKeeper) HasCredential(ctx sdk.Context, id string) bool {
	_, ok := m.Credentials[id]
	return ok
}

// DeleteCredential deletes a credential (test helper)
func (m *MockVCRegistryKeeper) DeleteCredential(ctx sdk.Context, id string) {
	delete(m.Credentials, id)
}

// ResetMock resets all mock data (test helper)
func (m *MockVCRegistryKeeper) ResetMock() {
	m.Credentials = make(map[string]interface{})
	m.ConfidenceScores = make(map[string]int64)
	m.IRScores = make(map[string]uint64)
	m.VerificationStatus = make(map[string]bool)
}
