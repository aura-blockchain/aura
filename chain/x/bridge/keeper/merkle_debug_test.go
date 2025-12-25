// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/aequitas/aura/chain/x/bridge/keeper"
)

// TestDebugMerkleProof helps understand the proof structure
func TestDebugMerkleProof(t *testing.T) {
	transactions := [][]byte{
		[]byte("transaction1"),
		[]byte("transaction2"),
		[]byte("transaction3"),
		[]byte("transaction4"),
	}

	// Compute root
	root := keeper.ComputeMerkleRoot(transactions)
	fmt.Printf("\nMerkle Root: %s\n", hex.EncodeToString(root))

	// Generate proof for index 1
	proof, err := keeper.GenerateMerkleProof(transactions, 1)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	fmt.Printf("\nProof for transaction at index 1:\n")
	fmt.Printf("  Leaf: %s\n", hex.EncodeToString(proof.Leaf))
	fmt.Printf("  Root: %s\n", hex.EncodeToString(proof.Root))
	fmt.Printf("  Proof siblings (%d):\n", len(proof.Proof))
	for i, sibling := range proof.Proof {
		fmt.Printf("    [%d] Index=%d Hash=%s\n", i, proof.Indices[i], hex.EncodeToString(sibling))
	}

	// Manually trace verification
	fmt.Printf("\nManual verification trace:\n")
	currentHash := proof.Leaf
	fmt.Printf("  Start: %s\n", hex.EncodeToString(currentHash))

	for i := 0; i < len(proof.Proof); i++ {
		sibling := proof.Proof[i]
		idx := proof.Indices[i]

		var combined []byte
		if idx%2 == 1 {
			// Sibling is on the right
			combined = append(currentHash, sibling...)
			fmt.Printf("  Level %d: current || sibling (idx=%d)\n", i, idx)
		} else {
			// Sibling is on the left
			combined = append(sibling, currentHash...)
			fmt.Printf("  Level %d: sibling || current (idx=%d)\n", i, idx)
		}

		hash := sha256.Sum256(combined)
		currentHash = hash[:]
		fmt.Printf("    -> %s\n", hex.EncodeToString(currentHash))
	}

	fmt.Printf("  Final: %s\n", hex.EncodeToString(currentHash))
	fmt.Printf("  Matches root: %v\n", hex.EncodeToString(currentHash) == hex.EncodeToString(root))

	// Test with all 4 indices
	fmt.Printf("\n\nTesting all transaction indices:\n")
	for idx := 0; idx < len(transactions); idx++ {
		proof, err := keeper.GenerateMerkleProof(transactions, idx)
		if err != nil {
			fmt.Printf("Index %d: ERROR - %v\n", idx, err)
			continue
		}

		valid := keeper.VerifyMerkleProof(proof)
		fmt.Printf("Index %d: Valid=%v, ProofLen=%d\n", idx, valid, len(proof.Proof))
	}
}
