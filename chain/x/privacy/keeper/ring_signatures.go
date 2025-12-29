// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/privacy/types"
)

// RingSignature represents a ring signature
type RingSignature struct {
	PublicKeys [][]byte
	Signature  []byte
	KeyImage   []byte
	Message    []byte
}

// VerifyRingSignature verifies a ring signature
func (k Keeper) VerifyRingSignature(ctx context.Context, signature *RingSignature) (bool, error) {
	params := k.GetParams(ctx)

	if !params.EnableRingSignatures {
		return false, fmt.Errorf("ring signatures not enabled")
	}

	// Validate ring size
	ringSize := len(signature.PublicKeys)
	if ringSize < int(params.MinRingSize) {
		return false, fmt.Errorf("ring size %d below minimum %d", ringSize, params.MinRingSize)
	}
	if ringSize > int(params.MaxRingSize) {
		return false, fmt.Errorf("ring size %d exceeds maximum %d", ringSize, params.MaxRingSize)
	}

	// Verify key image hasn't been used (prevent double-spend)
	if k.KeyImageExists(ctx, signature.KeyImage) {
		return false, types.ErrKeyImageAlreadyUsed
	}

	// Simplified ring signature verification
	// In production, use proper cryptographic verification (e.g., Monero-style LSAG)
	isValid := k.verifyRingSignatureCrypto(signature)
	if !isValid {
		return false, fmt.Errorf("invalid ring signature")
	}

	// Store key image to prevent reuse
	if err := k.StoreKeyImage(ctx, signature.KeyImage); err != nil {
		return false, err
	}

	return true, nil
}

// verifyRingSignatureCrypto performs cryptographic LSAG (Linkable Spontaneous Anonymous Group)
// ring signature verification using the secp256k1 curve.
//
// LSAG Verification Algorithm:
// 1. Parse signature components: c_0, s_0, s_1, ..., s_{n-1}
// 2. For each ring member i:
//   - Compute L_i = s_i * G + c_i * P_i
//   - Compute R_i = s_i * H_p(P_i) + c_i * I  (where I is key image)
//   - Compute c_{i+1} = H(m || L_i || R_i)
//
// 3. Verify c_n == c_0 (ring closes)
func (k Keeper) verifyRingSignatureCrypto(signature *RingSignature) bool {
	if len(signature.Signature) == 0 || len(signature.KeyImage) == 0 || len(signature.PublicKeys) == 0 {
		return false
	}

	curve := elliptic.P256()
	ringSize := len(signature.PublicKeys)

	// Signature format: c_0 (32 bytes) + s_0...s_{n-1} (32 bytes each)
	expectedSigLen := 32 + (ringSize * 32)
	if len(signature.Signature) != expectedSigLen {
		return false
	}

	// Parse c_0 (initial challenge)
	c0 := new(big.Int).SetBytes(signature.Signature[:32])

	// Parse s values
	sValues := make([]*big.Int, ringSize)
	for i := 0; i < ringSize; i++ {
		offset := 32 + (i * 32)
		sValues[i] = new(big.Int).SetBytes(signature.Signature[offset : offset+32])
	}

	// Parse key image I (must be a valid curve point)
	if len(signature.KeyImage) != 65 {
		return false
	}
	Ix, Iy := elliptic.Unmarshal(curve, signature.KeyImage)
	if Ix == nil || Iy == nil {
		return false
	}

	// Verify each ring member and compute challenges
	c := c0
	for i := 0; i < ringSize; i++ {
		// Parse public key P_i
		if len(signature.PublicKeys[i]) != 65 {
			return false
		}
		Px, Py := elliptic.Unmarshal(curve, signature.PublicKeys[i])
		if Px == nil || Py == nil {
			return false
		}

		// Compute L_i = s_i * G + c_i * P_i
		sGx, sGy := curve.ScalarBaseMult(sValues[i].Bytes())
		cPx, cPy := curve.ScalarMult(Px, Py, c.Bytes())
		Lx, Ly := curve.Add(sGx, sGy, cPx, cPy)

		// Compute H_p(P_i) - hash to curve point
		Hx, Hy := hashToPoint(curve, signature.PublicKeys[i])

		// Compute R_i = s_i * H_p(P_i) + c_i * I
		sHx, sHy := curve.ScalarMult(Hx, Hy, sValues[i].Bytes())
		cIx, cIy := curve.ScalarMult(Ix, Iy, c.Bytes())
		Rx, Ry := curve.Add(sHx, sHy, cIx, cIy)

		// Compute c_{i+1} = H(m || L_i || R_i)
		c = computeChallenge(curve, signature.Message, Lx, Ly, Rx, Ry)
	}

	// Ring closes if c_n == c_0
	return c.Cmp(c0) == 0
}

// hashToPoint deterministically hashes data to a curve point using try-and-increment
func hashToPoint(curve elliptic.Curve, data []byte) (*big.Int, *big.Int) {
	for i := uint32(0); i < 256; i++ {
		h := sha256.New()
		h.Write(data)
		_ = binary.Write(h, binary.BigEndian, i)
		xBytes := h.Sum(nil)

		x := new(big.Int).SetBytes(xBytes)
		x.Mod(x, curve.Params().P)

		// Try to find y for this x
		y := computeY(curve, x)
		if y != nil {
			return x, y
		}
	}
	// Fallback: use generator point (should not happen with valid input)
	return curve.Params().Gx, curve.Params().Gy
}

// computeY computes y coordinate from x on the curve (y^2 = x^3 - 3x + b mod p)
func computeY(curve elliptic.Curve, x *big.Int) *big.Int {
	params := curve.Params()
	p := params.P

	// y^2 = x^3 - 3x + b mod p
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Mod(x3, p)

	threeX := new(big.Int).Mul(x, big.NewInt(3))
	threeX.Mod(threeX, p)

	y2 := new(big.Int).Sub(x3, threeX)
	y2.Add(y2, params.B)
	y2.Mod(y2, p)

	// Compute square root using Tonelli-Shanks (p ≡ 3 mod 4 for P256)
	exp := new(big.Int).Add(p, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	y := new(big.Int).Exp(y2, exp, p)

	// Verify y^2 == y2 mod p
	check := new(big.Int).Mul(y, y)
	check.Mod(check, p)
	if check.Cmp(y2) != 0 {
		return nil
	}
	return y
}

// computeChallenge computes H(m || L || R) mod n
func computeChallenge(curve elliptic.Curve, message []byte, Lx, Ly, Rx, Ry *big.Int) *big.Int {
	h := sha256.New()
	h.Write(message)
	h.Write(elliptic.Marshal(curve, Lx, Ly))
	h.Write(elliptic.Marshal(curve, Rx, Ry))
	hash := h.Sum(nil)

	c := new(big.Int).SetBytes(hash)
	c.Mod(c, curve.Params().N)
	return c
}

// KeyImageExists checks if a key image has been used
func (k Keeper) KeyImageExists(ctx context.Context, keyImage []byte) bool {
	store := k.getStore(ctx)
	key := append(types.KeyImagePrefix, keyImage...)
	return store.Has(key)
}

// StoreKeyImage stores a key image to prevent double-spending
func (k Keeper) StoreKeyImage(ctx context.Context, keyImage []byte) error {
	store := k.getStore(ctx)
	key := append(types.KeyImagePrefix, keyImage...)

	if store.Has(key) {
		return types.ErrKeyImageAlreadyUsed
	}

	store.Set(key, []byte{0x01})
	return nil
}

// GenerateRingSignature generates a valid LSAG ring signature.
// secretKey is the private key corresponding to publicKeys[secretIndex].
func (k Keeper) GenerateRingSignature(ctx context.Context, message []byte, publicKeys [][]byte, secretKey *ecdsa.PrivateKey, secretIndex int) (*RingSignature, error) {
	curve := elliptic.P256()
	ringSize := len(publicKeys)

	if secretIndex < 0 || secretIndex >= ringSize {
		return nil, fmt.Errorf("secret index out of range")
	}

	// Validate all public keys are valid curve points
	for i, pk := range publicKeys {
		if len(pk) != 65 {
			return nil, fmt.Errorf("invalid public key length at index %d", i)
		}
		x, y := elliptic.Unmarshal(curve, pk)
		if x == nil || y == nil {
			return nil, fmt.Errorf("invalid public key at index %d", i)
		}
	}

	// Generate key image: I = x * H_p(P) where x is secret key, P is public key
	Hx, Hy := hashToPoint(curve, publicKeys[secretIndex])
	Ix, Iy := curve.ScalarMult(Hx, Hy, secretKey.D.Bytes())
	keyImage := elliptic.Marshal(curve, Ix, Iy)

	// Generate random scalars for non-signer indices and compute challenges
	sValues := make([]*big.Int, ringSize)
	cValues := make([]*big.Int, ringSize)

	// Start from secretIndex + 1
	// Generate random k for the signer
	k_rand, err := ecdsa.GenerateKey(curve, newDeterministicReader(message, secretKey.D.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random k: %w", err)
	}
	alpha := k_rand.D

	// Compute L = alpha * G and R = alpha * H_p(P_secretIndex)
	Lx, Ly := curve.ScalarBaseMult(alpha.Bytes())
	Rx, Ry := curve.ScalarMult(Hx, Hy, alpha.Bytes())

	// Compute c_{secretIndex+1}
	nextIndex := (secretIndex + 1) % ringSize
	cValues[nextIndex] = computeChallenge(curve, message, Lx, Ly, Rx, Ry)

	// Forward pass: compute c and random s for all other ring members
	var lastLx, lastLy, lastRx, lastRy *big.Int
	for i := 1; i < ringSize; i++ {
		idx := (secretIndex + i) % ringSize
		nextIdx := (idx + 1) % ringSize

		// Generate random s_i
		sRand, _ := ecdsa.GenerateKey(curve, newDeterministicReader(message, append(alpha.Bytes(), byte(idx))))
		sValues[idx] = sRand.D

		// Parse P_idx
		Px, Py := elliptic.Unmarshal(curve, publicKeys[idx])

		// Compute L_i = s_i * G + c_i * P_i
		sGx, sGy := curve.ScalarBaseMult(sValues[idx].Bytes())
		cPx, cPy := curve.ScalarMult(Px, Py, cValues[idx].Bytes())
		Li_x, Li_y := curve.Add(sGx, sGy, cPx, cPy)

		// Compute R_i = s_i * H_p(P_i) + c_i * I
		Hi_x, Hi_y := hashToPoint(curve, publicKeys[idx])
		sHx, sHy := curve.ScalarMult(Hi_x, Hi_y, sValues[idx].Bytes())
		cIx, cIy := curve.ScalarMult(Ix, Iy, cValues[idx].Bytes())
		Ri_x, Ri_y := curve.Add(sHx, sHy, cIx, cIy)

		// Store for computing c_secretIndex from the last iteration
		lastLx, lastLy, lastRx, lastRy = Li_x, Li_y, Ri_x, Ri_y

		// c_{i+1} = H(m || L_i || R_i)
		if nextIdx != secretIndex {
			cValues[nextIdx] = computeChallenge(curve, message, Li_x, Li_y, Ri_x, Ri_y)
		}
	}

	// Compute c_secretIndex from the last ring member
	if ringSize > 1 {
		cValues[secretIndex] = computeChallenge(curve, message, lastLx, lastLy, lastRx, lastRy)
	} else {
		// Single member ring: c_0 comes from initial L, R
		cValues[secretIndex] = computeChallenge(curve, message, Lx, Ly, Rx, Ry)
	}

	// Compute s_secretIndex = alpha - c_secretIndex * x mod n
	cx := new(big.Int).Mul(cValues[secretIndex], secretKey.D)
	cx.Mod(cx, curve.Params().N)
	sValues[secretIndex] = new(big.Int).Sub(alpha, cx)
	if sValues[secretIndex].Sign() < 0 {
		sValues[secretIndex].Add(sValues[secretIndex], curve.Params().N)
	}
	sValues[secretIndex].Mod(sValues[secretIndex], curve.Params().N)

	// Build signature: c_0 || s_0 || s_1 || ... || s_{n-1}
	sig := make([]byte, 32+(ringSize*32))
	copy(sig[:32], padTo32(cValues[0]))
	for i := 0; i < ringSize; i++ {
		copy(sig[32+(i*32):32+((i+1)*32)], padTo32(sValues[i]))
	}

	return &RingSignature{
		PublicKeys: publicKeys,
		Signature:  sig,
		KeyImage:   keyImage,
		Message:    message,
	}, nil
}

// padTo32 pads a big.Int to 32 bytes
func padTo32(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) >= 32 {
		return b[:32]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

// deterministicReader provides deterministic randomness for signature generation
type deterministicReader struct {
	seed []byte
	idx  int
}

func newDeterministicReader(message, secret []byte) *deterministicReader {
	h := sha256.New()
	h.Write(message)
	h.Write(secret)
	return &deterministicReader{seed: h.Sum(nil), idx: 0}
}

func (r *deterministicReader) Read(p []byte) (n int, err error) {
	for i := range p {
		h := sha256.New()
		h.Write(r.seed)
		_ = binary.Write(h, binary.BigEndian, uint64(r.idx))
		hash := h.Sum(nil)
		p[i] = hash[0]
		r.idx++
	}
	return len(p), nil
}

// generateKeyImage generates a key image I = x * H_p(P) for a given private key
func (k Keeper) generateKeyImage(privateKey *ecdsa.PrivateKey) []byte {
	curve := elliptic.P256()
	publicKey := elliptic.Marshal(curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	Hx, Hy := hashToPoint(curve, publicKey)
	Ix, Iy := curve.ScalarMult(Hx, Hy, privateKey.D.Bytes())
	return elliptic.Marshal(curve, Ix, Iy)
}

// VerifyLinkableRingSignature verifies a linkable ring signature (LSAG)
func (k Keeper) VerifyLinkableRingSignature(ctx context.Context, signature *RingSignature) (bool, error) {
	// Linkable ring signatures allow detection of double-signing by same signer
	// while maintaining anonymity within the ring

	// Check if key image already used
	if k.KeyImageExists(ctx, signature.KeyImage) {
		return false, types.ErrKeyImageAlreadyUsed
	}

	// Verify the ring signature
	valid, err := k.VerifyRingSignature(ctx, signature)
	if err != nil {
		return false, err
	}

	return valid, nil
}

// GetRingMembers retrieves potential ring members from the decoy pool for a ring signature.
// The decoy pool contains previously registered public keys that can be used to obfuscate
// the true signer. Members are selected pseudo-randomly based on block entropy.
func (k Keeper) GetRingMembers(ctx context.Context, ringSize int) ([][]byte, error) {
	params := k.GetParams(ctx)

	if ringSize < int(params.MinRingSize) || ringSize > int(params.MaxRingSize) {
		return nil, fmt.Errorf("invalid ring size: %d", ringSize)
	}

	// Collect all available ring members from the decoy pool
	store := k.getStore(ctx)
	iterator := store.Iterator(types.RingMemberPrefix, storeprefixend(types.RingMemberPrefix))
	defer iterator.Close()

	var allMembers [][]byte
	for ; iterator.Valid(); iterator.Next() {
		pubKey := iterator.Value()
		// Validate the public key is a valid secp256k1/P256 point (65 bytes uncompressed)
		if len(pubKey) == 65 || len(pubKey) == 33 {
			allMembers = append(allMembers, pubKey)
		}
	}

	// If not enough members in the pool, generate synthetic ones using deterministic randomness
	// This ensures ring signatures can still work when the pool is sparse
	if len(allMembers) < ringSize {
		needed := ringSize - len(allMembers)
		syntheticMembers := k.generateSyntheticRingMembers(ctx, needed)
		allMembers = append(allMembers, syntheticMembers...)
	}

	// Select ringSize members pseudo-randomly using block entropy
	// This provides unpredictable but deterministic selection
	selected, err := k.selectRandomMembers(ctx, allMembers, ringSize)
	if err != nil {
		return nil, fmt.Errorf("failed to select ring members: %w", err)
	}

	return selected, nil
}

// generateSyntheticRingMembers creates deterministic synthetic public keys when
// the decoy pool has insufficient members. Uses block hash + index for determinism.
func (k Keeper) generateSyntheticRingMembers(ctx context.Context, count int) [][]byte {
	curve := elliptic.P256()
	members := make([][]byte, count)

	// Use block data for deterministic generation
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHash := sdkCtx.HeaderHash()
	blockTime := sdkCtx.BlockTime().UnixNano()

	for i := 0; i < count; i++ {
		// Create deterministic seed from block data and index
		h := sha256.New()
		h.Write(blockHash)
		h.Write([]byte("synthetic_ring_member"))
		_ = binary.Write(h, binary.BigEndian, int64(i))
		_ = binary.Write(h, binary.BigEndian, blockTime)
		seed := h.Sum(nil)

		// Derive a valid curve point using try-and-increment
		for attempt := uint32(0); attempt < 256; attempt++ {
			attemptHash := sha256.New()
			attemptHash.Write(seed)
			_ = binary.Write(attemptHash, binary.BigEndian, attempt)
			xBytes := attemptHash.Sum(nil)

			x := new(big.Int).SetBytes(xBytes)
			x.Mod(x, curve.Params().P)

			// Try to find y for this x
			y := computeY(curve, x)
			if y != nil {
				members[i] = elliptic.Marshal(curve, x, y)
				break
			}
		}

		// Fallback if no valid point found (extremely unlikely)
		if members[i] == nil {
			members[i] = elliptic.Marshal(curve, curve.Params().Gx, curve.Params().Gy)
		}
	}

	return members
}

// selectRandomMembers selects n members from the pool using Fisher-Yates shuffle
// with entropy derived from block data.
func (k Keeper) selectRandomMembers(ctx context.Context, pool [][]byte, n int) ([][]byte, error) {
	if len(pool) < n {
		return nil, fmt.Errorf("pool size %d is less than requested %d", len(pool), n)
	}

	if len(pool) == n {
		return pool, nil
	}

	// Create a copy to shuffle
	shuffled := make([][]byte, len(pool))
	copy(shuffled, pool)

	// Use block entropy for deterministic random selection
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHash := sdkCtx.HeaderHash()

	// Fisher-Yates partial shuffle (only shuffle first n elements)
	for i := 0; i < n; i++ {
		// Generate random index using block hash and position
		h := sha256.New()
		h.Write(blockHash)
		h.Write([]byte("ring_member_select"))
		_ = binary.Write(h, binary.BigEndian, int64(i))
		hash := h.Sum(nil)

		// Use first 8 bytes as random value
		randVal := uint64(0)
		for j := 0; j < 8 && j < len(hash); j++ {
			randVal = (randVal << 8) | uint64(hash[j])
		}

		// Select random index from remaining pool
		remaining := len(shuffled) - i
		j := i + int(randVal%uint64(remaining))

		// Swap
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:n], nil
}

// storeprefixend returns the end key for a prefix iteration
func storeprefixend(prefix []byte) []byte {
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}
	return nil
}

// AddRingMember adds a public key to the ring member pool
func (k Keeper) AddRingMember(ctx context.Context, publicKey []byte) error {
	store := k.getStore(ctx)
	key := append(types.RingMemberPrefix, publicKey...)

	store.Set(key, publicKey)

	return nil
}

// RemoveRingMember removes a public key from the ring member pool
func (k Keeper) RemoveRingMember(ctx context.Context, publicKey []byte) error {
	store := k.getStore(ctx)
	key := append(types.RingMemberPrefix, publicKey...)

	store.Delete(key)

	return nil
}
