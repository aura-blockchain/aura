package keeper

import (
	"crypto/sha256"
	"errors"

	"github.com/aequitas/aura/chain/x/bridge/types"
)

// ComputeMerkleRoot computes the Merkle root of a list of data elements
func ComputeMerkleRoot(data [][]byte) []byte {
	if len(data) == 0 {
		return nil
	}

	if len(data) == 1 {
		hash := sha256.Sum256(data[0])
		return hash[:]
	}

	// Hash all leaf nodes
	var nodes [][]byte
	for _, d := range data {
		hash := sha256.Sum256(d)
		nodes = append(nodes, hash[:])
	}

	// Build tree bottom-up
	for len(nodes) > 1 {
		var nextLevel [][]byte

		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				// Combine two nodes
				combined := append(nodes[i], nodes[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			} else {
				// Odd number of nodes - duplicate the last one
				combined := append(nodes[i], nodes[i]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])
			}
		}

		nodes = nextLevel
	}

	return nodes[0]
}

// GenerateMerkleProof generates a Merkle proof for a specific index in the data
func GenerateMerkleProof(data [][]byte, index int) (*types.MerkleProof, error) {
	if index < 0 || index >= len(data) {
		return nil, errors.New("index out of range")
	}

	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	// Hash the leaf
	leafHash := sha256.Sum256(data[index])

	// Hash all leaf nodes
	var nodes [][]byte
	for _, d := range data {
		hash := sha256.Sum256(d)
		nodes = append(nodes, hash[:])
	}

	// Collect proof siblings
	var proofHashes [][]byte
	var indices []uint64
	currentIndex := uint64(index)

	// Build proof by traversing up the tree
	for len(nodes) > 1 {
		var nextLevel [][]byte
		levelIndex := currentIndex

		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				// Combine two nodes
				combined := append(nodes[i], nodes[i+1]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])

				// If this pair contains our node, add sibling to proof
				if i == int(levelIndex) || i+1 == int(levelIndex) {
					siblingIndex := i
					if i == int(levelIndex) {
						siblingIndex = i + 1
					}
					proofHashes = append(proofHashes, nodes[siblingIndex])
					indices = append(indices, uint64(siblingIndex))
				}
			} else {
				// Odd number of nodes - duplicate the last one
				combined := append(nodes[i], nodes[i]...)
				hash := sha256.Sum256(combined)
				nextLevel = append(nextLevel, hash[:])

				if i == int(levelIndex) {
					proofHashes = append(proofHashes, nodes[i])
					indices = append(indices, uint64(i))
				}
			}
		}

		nodes = nextLevel
		currentIndex = currentIndex / 2
	}

	root := nodes[0]

	return &types.MerkleProof{
		Root:    root,
		Leaf:    leafHash[:],
		Proof:   proofHashes,
		Indices: indices,
	}, nil
}

// VerifyMerkleProof verifies a Merkle proof
func VerifyMerkleProof(proof *types.MerkleProof) bool {
	if proof == nil {
		return false
	}

	// Special case: single element tree (no siblings needed)
	if len(proof.Proof) == 0 {
		// For a single element, the leaf hash should equal the root
		if len(proof.Leaf) != len(proof.Root) {
			return false
		}
		for i := range proof.Leaf {
			if proof.Leaf[i] != proof.Root[i] {
				return false
			}
		}
		return true
	}

	// Start with the leaf hash
	currentHash := proof.Leaf

	// Traverse up the tree
	for i := 0; i < len(proof.Proof); i++ {
		sibling := proof.Proof[i]

		// Combine with sibling
		var combined []byte
		if i < len(proof.Indices) && proof.Indices[i]%2 == 1 {
			// Sibling is on the right
			combined = append(currentHash, sibling...)
		} else {
			// Sibling is on the left
			combined = append(sibling, currentHash...)
		}

		hash := sha256.Sum256(combined)
		currentHash = hash[:]
	}

	// Check if final hash matches root
	if len(currentHash) != len(proof.Root) {
		return false
	}

	for i := range currentHash {
		if currentHash[i] != proof.Root[i] {
			return false
		}
	}

	return true
}
