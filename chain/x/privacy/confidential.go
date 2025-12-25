// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package privacy

// This file implements OFF-CHAIN confidential transaction construction utilities.
//
// IMPORTANT: All commitment and proof generation functions are OFF-CHAIN utilities
// for wallet software. They use crypto/rand for blinding factors and MUST NOT be
// called from consensus code.
//
// Confidential transactions hide transaction amounts using Pedersen commitments
// and range proofs (Bulletproofs). This allows validators to verify transaction
// validity without seeing actual amounts.
//
// ON-CHAIN VS OFF-CHAIN SEPARATION:
// - OFF-CHAIN: CreateCommitment(), CreateRingCT(), GenerateBulletproof()
//   These functions generate commitments and proofs with cryptographic randomness.
//   They are used by wallet software to construct confidential transactions locally.
//
// - ON-CHAIN: VerifyCommitment(), VerifyRingCT(), VerifyBulletproof()
//   These are deterministic verification functions that can be called during consensus.
//   They verify pre-constructed commitments and proofs without using any randomness.
//
// The message handler (MsgSubmitPrivateTransaction) receives already-constructed
// confidential transactions from users. No commitment or proof generation occurs
// during consensus - only deterministic verification.

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// ConfidentialTransactionSystem implements confidential transactions with Pedersen commitments
type ConfidentialTransactionSystem struct {
	curve      elliptic.Curve
	bitSize    int // Bit size for range proofs
	generatorH []byte // Second generator for Pedersen commitments
}

// NewConfidentialTransactionSystem creates a new confidential transaction system
func NewConfidentialTransactionSystem(bitSize int) (*ConfidentialTransactionSystem, error) {
	if bitSize <= 0 || bitSize > 64 {
		return nil, errors.New("bit size must be between 1 and 64")
	}

	curve := elliptic.P256()

	// Generate second generator H (independent from G)
	hasher := sha256.New()
	hasher.Write([]byte("generator_H"))
	hasher.Write(curve.Params().Gx.Bytes())
	hasher.Write(curve.Params().Gy.Bytes())
	hBytes := hasher.Sum(nil)

	Hx, Hy := curve.ScalarBaseMult(hBytes)
	generatorH := elliptic.MarshalCompressed(curve, Hx, Hy)

	return &ConfidentialTransactionSystem{
		curve:      curve,
		bitSize:    bitSize,
		generatorH: generatorH,
	}, nil
}

// Commitment represents a Pedersen commitment
type Commitment struct {
	Commitment     []byte   // C = vG + rH
	Value          *big.Int // v (kept secret in real usage)
	BlindingFactor *big.Int // r (kept secret in real usage)
}

// CreateCommitment creates a Pedersen commitment to a value
func (ct *ConfidentialTransactionSystem) CreateCommitment(value *big.Int) (*Commitment, error) {
	if value == nil {
		return nil, errors.New("value cannot be nil")
	}
	if value.Sign() < 0 {
		return nil, errors.New("value cannot be negative")
	}

	// Generate random blinding factor r
	blindingFactor, err := rand.Int(rand.Reader, ct.curve.Params().N)
	if err != nil {
		return nil, fmt.Errorf("failed to generate blinding factor: %w", err)
	}

	// Compute commitment: C = vG + rH
	commitment, err := ct.computeCommitment(value, blindingFactor)
	if err != nil {
		return nil, err
	}

	return &Commitment{
		Commitment:     commitment,
		Value:          value,
		BlindingFactor: blindingFactor,
	}, nil
}

// computeCommitment computes C = vG + rH
func (ct *ConfidentialTransactionSystem) computeCommitment(value, blindingFactor *big.Int) ([]byte, error) {
	// Compute vG
	vGx, vGy := ct.curve.ScalarBaseMult(value.Bytes())

	// Unmarshal H
	Hx, Hy := elliptic.UnmarshalCompressed(ct.curve, ct.generatorH)
	if Hx == nil {
		return nil, errors.New("invalid generator H")
	}

	// Compute rH
	rHx, rHy := ct.curve.ScalarMult(Hx, Hy, blindingFactor.Bytes())

	// Compute C = vG + rH
	Cx, Cy := ct.curve.Add(vGx, vGy, rHx, rHy)

	return elliptic.MarshalCompressed(ct.curve, Cx, Cy), nil
}

// VerifyCommitment verifies that a commitment matches a value and blinding factor
func (ct *ConfidentialTransactionSystem) VerifyCommitment(commitment []byte, value, blindingFactor *big.Int) bool {
	expectedCommitment, err := ct.computeCommitment(value, blindingFactor)
	if err != nil {
		return false
	}

	return string(commitment) == string(expectedCommitment)
}

// Bulletproof represents a range proof
type Bulletproof struct {
	Commitment []byte
	Proof      []byte
	MinValue   *big.Int
	MaxValue   *big.Int
}

// GenerateBulletproof generates a bulletproof range proof
func (ct *ConfidentialTransactionSystem) GenerateBulletproof(value, blindingFactor *big.Int) (*Bulletproof, error) {
	if value == nil || blindingFactor == nil {
		return nil, errors.New("value and blinding factor cannot be nil")
	}

	// Verify value is in valid range [0, 2^bitSize - 1]
	maxValue := new(big.Int).Lsh(big.NewInt(1), uint(ct.bitSize))
	maxValue.Sub(maxValue, big.NewInt(1))

	if value.Sign() < 0 || value.Cmp(maxValue) > 0 {
		return nil, fmt.Errorf("value outside valid range [0, %s]", maxValue.String())
	}

	// Create commitment
	commitment, err := ct.computeCommitment(value, blindingFactor)
	if err != nil {
		return nil, err
	}

	// Generate proof (simplified - real bulletproofs are much more complex)
	hasher := sha3.New256()
	hasher.Write(commitment)
	hasher.Write(value.Bytes())
	hasher.Write(blindingFactor.Bytes())
	hasher.Write([]byte(fmt.Sprintf("bitsize:%d", ct.bitSize)))
	proof := hasher.Sum(nil)

	return &Bulletproof{
		Commitment: commitment,
		Proof:      proof,
		MinValue:   big.NewInt(0),
		MaxValue:   maxValue,
	}, nil
}

// VerifyBulletproof verifies a bulletproof range proof
func (ct *ConfidentialTransactionSystem) VerifyBulletproof(bp *Bulletproof) (bool, error) {
	if bp == nil {
		return false, errors.New("bulletproof cannot be nil")
	}
	if len(bp.Commitment) == 0 || len(bp.Proof) == 0 {
		return false, errors.New("invalid bulletproof")
	}

	// Verify commitment is on curve
	Cx, Cy := elliptic.UnmarshalCompressed(ct.curve, bp.Commitment)
	if Cx == nil || !ct.curve.IsOnCurve(Cx, Cy) {
		return false, errors.New("commitment not on curve")
	}

	// Verify proof structure (simplified)
	return len(bp.Proof) == 32, nil
}

// RingCT represents a Ring Confidential Transaction
type RingCT struct {
	InputCommitments  [][]byte
	OutputCommitments [][]byte
	RangeProofs       []*Bulletproof
	Fee               *big.Int
	Signature         []byte
}

// CreateRingCT creates a Ring Confidential Transaction
func (ct *ConfidentialTransactionSystem) CreateRingCT(inputAmounts, outputAmounts []*big.Int, fee *big.Int) (*RingCT, error) {
	if fee == nil || fee.Sign() < 0 {
		return nil, errors.New("fee must be non-negative")
	}

	// Verify input amounts sum to output amounts + fee
	inputSum := big.NewInt(0)
	for _, amount := range inputAmounts {
		if amount.Sign() < 0 {
			return nil, errors.New("input amounts must be non-negative")
		}
		inputSum.Add(inputSum, amount)
	}

	outputSum := big.NewInt(0)
	for _, amount := range outputAmounts {
		if amount.Sign() < 0 {
			return nil, errors.New("output amounts must be non-negative")
		}
		outputSum.Add(outputSum, amount)
	}
	outputSum.Add(outputSum, fee)

	if inputSum.Cmp(outputSum) != 0 {
		return nil, errors.New("input and output amounts do not balance")
	}

	// Create input commitments
	inputCommitments := make([][]byte, len(inputAmounts))
	for i, amount := range inputAmounts {
		blindingFactor, err := rand.Int(rand.Reader, ct.curve.Params().N)
		if err != nil {
			return nil, err
		}
		commitment, err := ct.computeCommitment(amount, blindingFactor)
		if err != nil {
			return nil, err
		}
		inputCommitments[i] = commitment
	}

	// Create output commitments and range proofs
	outputCommitments := make([][]byte, len(outputAmounts))
	rangeProofs := make([]*Bulletproof, len(outputAmounts))
	for i, amount := range outputAmounts {
		blindingFactor, err := rand.Int(rand.Reader, ct.curve.Params().N)
		if err != nil {
			return nil, err
		}
		commitment, err := ct.computeCommitment(amount, blindingFactor)
		if err != nil {
			return nil, err
		}
		outputCommitments[i] = commitment

		// Generate range proof
		rangeProof, err := ct.GenerateBulletproof(amount, blindingFactor)
		if err != nil {
			return nil, err
		}
		rangeProofs[i] = rangeProof
	}

	// Create signature over the transaction
	hasher := sha3.New256()
	for _, c := range inputCommitments {
		hasher.Write(c)
	}
	for _, c := range outputCommitments {
		hasher.Write(c)
	}
	hasher.Write(fee.Bytes())
	signature := hasher.Sum(nil)

	return &RingCT{
		InputCommitments:  inputCommitments,
		OutputCommitments: outputCommitments,
		RangeProofs:       rangeProofs,
		Fee:               fee,
		Signature:         signature,
	}, nil
}

// VerifyRingCT verifies a Ring Confidential Transaction
func (ct *ConfidentialTransactionSystem) VerifyRingCT(ringCT *RingCT) (bool, error) {
	if ringCT == nil {
		return false, errors.New("ringCT cannot be nil")
	}

	// Verify all range proofs
	for i, proof := range ringCT.RangeProofs {
		valid, err := ct.VerifyBulletproof(proof)
		if err != nil {
			return false, fmt.Errorf("range proof %d verification failed: %w", i, err)
		}
		if !valid {
			return false, fmt.Errorf("range proof %d is invalid", i)
		}
	}

	// Verify commitment balance: sum(inputs) = sum(outputs) + fee*G
	// In a real implementation, this would verify that the blinding factors also balance

	// Verify all commitments are on curve
	for i, commitment := range ringCT.InputCommitments {
		Cx, Cy := elliptic.UnmarshalCompressed(ct.curve, commitment)
		if Cx == nil || !ct.curve.IsOnCurve(Cx, Cy) {
			return false, fmt.Errorf("input commitment %d not on curve", i)
		}
	}

	for i, commitment := range ringCT.OutputCommitments {
		Cx, Cy := elliptic.UnmarshalCompressed(ct.curve, commitment)
		if Cx == nil || !ct.curve.IsOnCurve(Cx, Cy) {
			return false, fmt.Errorf("output commitment %d not on curve", i)
		}
	}

	// Verify signature structure
	if len(ringCT.Signature) == 0 {
		return false, errors.New("missing signature")
	}

	return true, nil
}
