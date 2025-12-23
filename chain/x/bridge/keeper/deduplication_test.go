package keeper

import (
	"testing"
)

// TestDeduplication_DeriveExternalAddressFromPubKey tests that the common
// deriveExternalAddressFromPubKey function produces consistent results.
// This test verifies the deduplication refactoring is correct.
func TestDeduplication_DeriveExternalAddressFromPubKey(t *testing.T) {
	// Create a minimal keeper for this test
	k := &Keeper{}

	// Generate a test public key (33 bytes compressed)
	testPubKey := make([]byte, 33)
	testPubKey[0] = 0x02 // Compressed public key prefix
	for i := 1; i < 33; i++ {
		testPubKey[i] = byte(i)
	}

	// Test that the common function produces the same result for both chains
	addressPaw := k.deriveExternalAddressFromPubKey(testPubKey, "paw")
	addressXai := k.deriveExternalAddressFromPubKey(testPubKey, "xai")

	// Since both chains use the same derivation algorithm, they should produce the same address
	if addressPaw != addressXai {
		t.Errorf("Expected same address for both chains, got paw=%s, xai=%s", addressPaw, addressXai)
	}

	// Verify address is not empty
	if addressPaw == "" {
		t.Error("Expected non-empty address")
	}

	// Test that the wrapper functions produce the same result as the common function
	addressPawWrapper := k.derivePawAddressFromPubKey(testPubKey)
	addressXaiWrapper := k.deriveXaiAddressFromPubKey(testPubKey)

	if addressPawWrapper != addressPaw {
		t.Errorf("PAW wrapper produced different address than direct call: %s != %s", addressPawWrapper, addressPaw)
	}

	if addressXaiWrapper != addressXai {
		t.Errorf("XAI wrapper produced different address than direct call: %s != %s", addressXaiWrapper, addressXai)
	}

	// Test with invalid public key length
	invalidPubKey := make([]byte, 32) // Wrong length
	addressInvalid := k.deriveExternalAddressFromPubKey(invalidPubKey, "paw")
	if addressInvalid != "" {
		t.Errorf("Expected empty string for invalid public key, got: %s", addressInvalid)
	}

	// Also test wrappers with invalid key
	addressInvalidPaw := k.derivePawAddressFromPubKey(invalidPubKey)
	addressInvalidXai := k.deriveXaiAddressFromPubKey(invalidPubKey)
	if addressInvalidPaw != "" || addressInvalidXai != "" {
		t.Error("Expected empty string for invalid public key in wrappers")
	}
}
