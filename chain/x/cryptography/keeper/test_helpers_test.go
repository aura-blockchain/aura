// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// GenerateDummyQuantumPublicKey generates a dummy public key of the correct size for testing.
// This is NOT cryptographically secure - it's only for testing purposes.
//
// In production, clients must generate real quantum-resistant keys using proper libraries:
// - CRYSTALS-Dilithium: Use pq-crystals/dilithium
// - CRYSTALS-Kyber: Use pq-crystals/kyber
// - Falcon: Use falcon-crypto/falcon
// - SPHINCS+: Use sphincs/sphincsplus
// - NTRU: Use NTRUEncrypt
func GenerateDummyQuantumPublicKey(algo cryptoproto.QuantumResistantAlgorithm) []byte {
	var size int
	switch algo {
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_DILITHIUM:
		size = 1312 // CRYSTALS-Dilithium2
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_CRYSTALS_KYBER:
		size = 800 // CRYSTALS-Kyber512
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_FALCON:
		size = 897 // Falcon-512
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_SPHINCS_PLUS:
		size = 32 // SPHINCS+-128s
	case cryptoproto.QuantumResistantAlgorithm_QUANTUM_RESISTANT_ALGORITHM_NTRU:
		size = 1230 // NTRU-HPS-2048-509
	default:
		return []byte{}
	}

	// Fill with deterministic dummy data (not cryptographically secure, just for testing)
	pubKey := make([]byte, size)
	for i := range pubKey {
		pubKey[i] = byte((i + int(algo)) % 256)
	}
	return pubKey
}
