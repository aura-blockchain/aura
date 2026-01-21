// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package privacy

// This file implements OFF-CHAIN ring signature utilities for privacy-preserving signing.
//
// IMPORTANT: All signature generation functions are OFF-CHAIN utilities that use
// crypto/rand for cryptographic randomness. NEVER call from consensus code.
//
// Ring signatures allow a signer to prove membership in a group (ring) without
// revealing which member actually signed. This provides anonymity within the ring.
//
// MLSAG (Multilayered Linkable Spontaneous Anonymous Group) signatures extend
// ring signatures to multiple keys per participant, as used in Monero.
//
// ON-CHAIN VS OFF-CHAIN SEPARATION:
// - OFF-CHAIN: Sign(), SignMLSAG()
//   These functions generate ring signatures using cryptographic randomness.
//   They require random values for the zero-knowledge property.
//   Used by wallet software to sign transactions locally.
//
// - ON-CHAIN: Verify(), VerifyMLSAG()
//   These are deterministic verification functions that can be called during consensus.
//   They verify ring signatures without using any randomness.
//
// Message handlers only verify pre-constructed ring signatures. No signature
// generation occurs during consensus.

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// RingSigner implements ring signature generation and verification
type RingSigner struct {
	curve elliptic.Curve
}

// NewRingSigner creates a new ring signature instance
func NewRingSigner() *RingSigner {
	return &RingSigner{
		curve: elliptic.P256(),
	}
}

// RingSignature represents a ring signature
type RingSignature struct {
	PublicKeys [][]byte   // Ring of public keys
	C          *big.Int   // Challenge value
	S          []*big.Int // Response values
	KeyImage   []byte     // Key image to prevent double-spending
}

// Sign creates a ring signature
func (rs *RingSigner) Sign(signerIndex int, privateKey *big.Int, publicKeys [][]byte, message []byte) (*RingSignature, error) {
	ringSize := len(publicKeys)
	if ringSize < 2 {
		return nil, errors.New("ring size must be at least 2")
	}
	if signerIndex < 0 || signerIndex >= ringSize {
		return nil, errors.New("invalid signer index")
	}

	// Generate key image: I = x*H(P) where x is private key, P is public key
	keyImage, err := rs.generateKeyImage(privateKey, publicKeys[signerIndex])
	if err != nil {
		return nil, fmt.Errorf("failed to generate key image: %w", err)
	}

	// Initialize random values
	s := make([]*big.Int, ringSize)
	for i := 0; i < ringSize; i++ {
		if i != signerIndex {
			s[i], err = rand.Int(rand.Reader, rs.curve.Params().N)
			if err != nil {
				return nil, err
			}
		}
	}

	// Generate random alpha for signer's index
	alpha, err := rand.Int(rand.Reader, rs.curve.Params().N)
	if err != nil {
		return nil, err
	}

	// Initialize c array and hasher
	c := make([]*big.Int, ringSize)
	hasher := sha3.New256()

	// Compute L_signer = alpha*G for signer's position
	Lx, Ly := rs.curve.ScalarBaseMult(alpha.Bytes())

	// Compute the initial challenge c_{signerIndex+1} from alpha*G
	hasher.Write(message)
	hasher.Write(elliptic.MarshalCompressed(rs.curve, Lx, Ly))
	cBytes := hasher.Sum(nil)

	// Set the challenge for the next position after signer
	nextIdx := (signerIndex + 1) % ringSize
	c[nextIdx] = new(big.Int).SetBytes(cBytes)
	c[nextIdx].Mod(c[nextIdx], rs.curve.Params().N)

	// Complete the ring: compute c values for all other positions
	// Start from signerIndex+2 and go around until we reach signerIndex
	for i := (signerIndex + 2) % ringSize; i != (signerIndex+1)%ringSize; i = (i + 1) % ringSize {
		prevIdx := (i - 1 + ringSize) % ringSize

		// Unmarshal public key for previous position
		Px, Py := elliptic.UnmarshalCompressed(rs.curve, publicKeys[prevIdx])
		if Px == nil {
			return nil, fmt.Errorf("invalid public key at index %d", prevIdx)
		}

		// Compute L_{prev} = s_{prev}*G + c_{prev}*P_{prev}
		sGx, sGy := rs.curve.ScalarBaseMult(s[prevIdx].Bytes())
		cPx, cPy := rs.curve.ScalarMult(Px, Py, c[prevIdx].Bytes())
		Lx, Ly = rs.curve.Add(sGx, sGy, cPx, cPy)

		// Hash to get c[i]
		hasher.Reset()
		hasher.Write(message)
		hasher.Write(elliptic.MarshalCompressed(rs.curve, Lx, Ly))
		cBytes = hasher.Sum(nil)
		c[i] = new(big.Int).SetBytes(cBytes)
		c[i].Mod(c[i], rs.curve.Params().N)
	}

	// Now compute c[signerIndex] from the last element before signer
	lastIdx := (signerIndex - 1 + ringSize) % ringSize
	Px, Py := elliptic.UnmarshalCompressed(rs.curve, publicKeys[lastIdx])
	if Px == nil {
		return nil, fmt.Errorf("invalid public key at index %d", lastIdx)
	}
	sGx, sGy := rs.curve.ScalarBaseMult(s[lastIdx].Bytes())
	cPx, cPy := rs.curve.ScalarMult(Px, Py, c[lastIdx].Bytes())
	Lx, Ly = rs.curve.Add(sGx, sGy, cPx, cPy)

	hasher.Reset()
	hasher.Write(message)
	hasher.Write(elliptic.MarshalCompressed(rs.curve, Lx, Ly))
	cBytes = hasher.Sum(nil)
	c[signerIndex] = new(big.Int).SetBytes(cBytes)
	c[signerIndex].Mod(c[signerIndex], rs.curve.Params().N)

	// Compute s for signer: s = alpha - c*x (mod n)
	cx := new(big.Int).Mul(c[signerIndex], privateKey)
	s[signerIndex] = new(big.Int).Sub(alpha, cx)
	s[signerIndex].Mod(s[signerIndex], rs.curve.Params().N)

	return &RingSignature{
		PublicKeys: publicKeys,
		C:          c[0],
		S:          s,
		KeyImage:   keyImage,
	}, nil
}

// Verify verifies a ring signature
func (rs *RingSigner) Verify(signature *RingSignature, message []byte) (bool, error) {
	if signature == nil {
		return false, errors.New("signature cannot be nil")
	}

	ringSize := len(signature.PublicKeys)
	if ringSize != len(signature.S) {
		return false, errors.New("signature components size mismatch")
	}

	c := signature.C
	for i := 0; i < ringSize; i++ {
		// Unmarshal public key
		Px, Py := elliptic.UnmarshalCompressed(rs.curve, signature.PublicKeys[i])
		if Px == nil {
			return false, fmt.Errorf("invalid public key at index %d", i)
		}

		// Compute L_i = s_i*G + c_i*P_i
		sGx, sGy := rs.curve.ScalarBaseMult(signature.S[i].Bytes())
		cPx, cPy := rs.curve.ScalarMult(Px, Py, c.Bytes())
		Lx, Ly := rs.curve.Add(sGx, sGy, cPx, cPy)

		// Hash to get next c
		hasher := sha3.New256()
		hasher.Write(message)
		hasher.Write(elliptic.MarshalCompressed(rs.curve, Lx, Ly))
		cBytes := hasher.Sum(nil)
		c = new(big.Int).SetBytes(cBytes)
		c.Mod(c, rs.curve.Params().N)
	}

	// Verify the ring closes: c should equal signature.C
	return c.Cmp(signature.C) == 0, nil
}

// generateKeyImage generates a key image from private key and public key
func (rs *RingSigner) generateKeyImage(privateKey *big.Int, publicKey []byte) ([]byte, error) {
	// Hash public key to get a point
	hasher := sha256.New()
	hasher.Write(publicKey)
	h := hasher.Sum(nil)

	// Use hash as scalar to generate point
	Hx, Hy := rs.curve.ScalarBaseMult(h)

	// Multiply by private key: I = x*H(P)
	Ix, Iy := rs.curve.ScalarMult(Hx, Hy, privateKey.Bytes())

	return elliptic.MarshalCompressed(rs.curve, Ix, Iy), nil
}

// MLSAGSigner implements Multilayered Linkable Spontaneous Anonymous Group signatures
type MLSAGSigner struct {
	curve elliptic.Curve
}

// NewMLSAGSigner creates a new MLSAG signer
func NewMLSAGSigner() *MLSAGSigner {
	return &MLSAGSigner{
		curve: elliptic.P256(),
	}
}

// MLSAGSignature represents an MLSAG signature
type MLSAGSignature struct {
	PublicKeyMatrix [][][]byte   // Matrix of public keys [ringSize][numKeys]
	C               *big.Int     // Initial challenge
	S               [][]*big.Int // Response matrix [ringSize][numKeys]
	KeyImages       [][]byte     // Key images for linkability
}

// SignMLSAG creates an MLSAG signature
func (ms *MLSAGSigner) SignMLSAG(signerIndex int, privateKeys []*big.Int, publicKeyMatrix [][][]byte, message []byte) (*MLSAGSignature, error) {
	ringSize := len(publicKeyMatrix)
	if ringSize < 2 {
		return nil, errors.New("ring size must be at least 2")
	}
	if signerIndex < 0 || signerIndex >= ringSize {
		return nil, errors.New("invalid signer index")
	}

	numKeys := len(privateKeys)
	if len(publicKeyMatrix[0]) != numKeys {
		return nil, errors.New("key count mismatch")
	}

	// Generate key images for each private key
	keyImages := make([][]byte, numKeys)
	for j := 0; j < numKeys; j++ {
		ki, err := ms.generateKeyImage(privateKeys[j], publicKeyMatrix[signerIndex][j])
		if err != nil {
			return nil, err
		}
		keyImages[j] = ki
	}

	// Initialize response matrix
	s := make([][]*big.Int, ringSize)
	for i := 0; i < ringSize; i++ {
		s[i] = make([]*big.Int, numKeys)
		if i != signerIndex {
			for j := 0; j < numKeys; j++ {
				var err error
				s[i][j], err = rand.Int(rand.Reader, ms.curve.Params().N)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// Generate random alpha values for signer
	alphas := make([]*big.Int, numKeys)
	for j := 0; j < numKeys; j++ {
		var err error
		alphas[j], err = rand.Int(rand.Reader, ms.curve.Params().N)
		if err != nil {
			return nil, err
		}
	}

	// Compute initial hash
	hasher := sha3.New256()
	hasher.Write(message)
	for j := 0; j < numKeys; j++ {
		Lx, Ly := ms.curve.ScalarBaseMult(alphas[j].Bytes())
		hasher.Write(elliptic.MarshalCompressed(ms.curve, Lx, Ly))
	}

	var c *big.Int
	cBytes := hasher.Sum(nil)
	c = new(big.Int).SetBytes(cBytes)
	c.Mod(c, ms.curve.Params().N)

	// Track the challenge value for position 0 (needed for signature.C)
	var c0 *big.Int

	// Complete the ring
	for i := (signerIndex + 1) % ringSize; i != signerIndex; i = (i + 1) % ringSize {
		hasher.Reset()
		hasher.Write(message)

		for j := 0; j < numKeys; j++ {
			Px, Py := elliptic.UnmarshalCompressed(ms.curve, publicKeyMatrix[i][j])
			if Px == nil {
				return nil, fmt.Errorf("invalid public key at [%d][%d]", i, j)
			}

			sGx, sGy := ms.curve.ScalarBaseMult(s[i][j].Bytes())
			cPx, cPy := ms.curve.ScalarMult(Px, Py, c.Bytes())
			Lx, Ly := ms.curve.Add(sGx, sGy, cPx, cPy)

			hasher.Write(elliptic.MarshalCompressed(ms.curve, Lx, Ly))
		}

		cBytes := hasher.Sum(nil)
		c = new(big.Int).SetBytes(cBytes)
		c.Mod(c, ms.curve.Params().N)

		// Save the challenge for position 0 when we compute it
		if (i+1)%ringSize == 0 {
			c0 = new(big.Int).Set(c)
		}
	}

	// Compute s for signer position
	s[signerIndex] = make([]*big.Int, numKeys)
	for j := 0; j < numKeys; j++ {
		cx := new(big.Int).Mul(c, privateKeys[j])
		s[signerIndex][j] = new(big.Int).Sub(alphas[j], cx)
		s[signerIndex][j].Mod(s[signerIndex][j], ms.curve.Params().N)
	}

	// Use c0 as the initial challenge for verification
	// If c0 wasn't set (signerIndex == 0), it means we never computed it in the loop
	// In that case, we need to compute it from the last position's challenge
	if c0 == nil {
		// This happens when signerIndex == 0
		// The loop would go: 1, 2, ..., ringSize-1
		// After the loop, c is the challenge for position 0
		c0 = c
	}

	return &MLSAGSignature{
		PublicKeyMatrix: publicKeyMatrix,
		C:               c0,
		S:               s,
		KeyImages:       keyImages,
	}, nil
}

// VerifyMLSAG verifies an MLSAG signature
func (ms *MLSAGSigner) VerifyMLSAG(signature *MLSAGSignature, message []byte) (bool, error) {
	if signature == nil {
		return false, errors.New("signature cannot be nil")
	}

	ringSize := len(signature.PublicKeyMatrix)
	numKeys := len(signature.PublicKeyMatrix[0])

	c := signature.C
	for i := 0; i < ringSize; i++ {
		hasher := sha3.New256()
		hasher.Write(message)

		for j := 0; j < numKeys; j++ {
			Px, Py := elliptic.UnmarshalCompressed(ms.curve, signature.PublicKeyMatrix[i][j])
			if Px == nil {
				return false, fmt.Errorf("invalid public key at [%d][%d]", i, j)
			}

			sGx, sGy := ms.curve.ScalarBaseMult(signature.S[i][j].Bytes())
			cPx, cPy := ms.curve.ScalarMult(Px, Py, c.Bytes())
			Lx, Ly := ms.curve.Add(sGx, sGy, cPx, cPy)

			hasher.Write(elliptic.MarshalCompressed(ms.curve, Lx, Ly))
		}

		cBytes := hasher.Sum(nil)
		c = new(big.Int).SetBytes(cBytes)
		c.Mod(c, ms.curve.Params().N)
	}

	return c.Cmp(signature.C) == 0, nil
}

// generateKeyImage generates a key image for MLSAG
func (ms *MLSAGSigner) generateKeyImage(privateKey *big.Int, publicKey []byte) ([]byte, error) {
	hasher := sha256.New()
	hasher.Write(publicKey)
	h := hasher.Sum(nil)

	Hx, Hy := ms.curve.ScalarBaseMult(h)
	Ix, Iy := ms.curve.ScalarMult(Hx, Hy, privateKey.Bytes())

	return elliptic.MarshalCompressed(ms.curve, Ix, Iy), nil
}
