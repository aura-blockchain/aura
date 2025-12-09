package v1beta1

import (
	"encoding/json"
)

// RawContractMessage is a raw JSON message passed to WASM contracts.
// This type provides type safety for contract messages while maintaining
// compatibility with the underlying bytes representation.
//
// Security considerations:
//   - Always validate JSON before passing to contracts
//   - Ensure proper sanitization of user inputs
//   - Consider size limits to prevent DoS attacks
type RawContractMessage []byte

// MarshalJSON implements json.Marshaler interface
func (m RawContractMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *RawContractMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return nil
	}
	*m = data
	return nil
}

// Bytes returns the underlying byte slice
func (m RawContractMessage) Bytes() []byte {
	return m
}

// ValidateBasic performs basic validation on the contract message
func (m RawContractMessage) ValidateBasic() error {
	if len(m) == 0 {
		return nil // Empty message is valid
	}

	// Verify it's valid JSON
	var js json.RawMessage
	if err := json.Unmarshal(m, &js); err != nil {
		return err
	}

	return nil
}
