// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

// BenchmarkCreateIdentity benchmarks DID string creation
func BenchmarkCreateIdentity(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("did:aura:test%d", i)
	}
}

// BenchmarkVerifyCredential benchmarks hash-based verification
func BenchmarkVerifyCredential(b *testing.B) {
	credID := "cred-1"
	did := "did:aura:test"
	data := []byte(credID + did)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sha256.Sum256(data)
	}
}

// BenchmarkZKProofVerification benchmarks proof data handling
func BenchmarkZKProofVerification(b *testing.B) {
	proof := make([]byte, 32)
	publicInputs := make([]byte, 32)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		combined := append(proof, publicInputs...)
		_ = sha256.Sum256(combined)
	}
}

// BenchmarkCreateSession benchmarks session ID generation
func BenchmarkCreateSession(b *testing.B) {
	userAddress := "aura1user"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sessionData := fmt.Sprintf("%s-%d", userAddress, i)
		_ = sha256.Sum256([]byte(sessionData))
	}
}

// BenchmarkVerifySignatureWithKey benchmarks signature hash computation
func BenchmarkVerifySignatureWithKey(b *testing.B) {
	message := []byte("test message")
	signature := make([]byte, 64)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hash := sha256.Sum256(message)
		copy(signature, hash[:])
	}
}

// BenchmarkVerifyPIICommitment benchmarks PII hashing
func BenchmarkVerifyPIICommitment(b *testing.B) {
	piiData := "name:John Doe,email:john@example.com"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sha256.Sum256([]byte(piiData))
	}
}

// BenchmarkCreateChangeRequest benchmarks request ID generation
func BenchmarkCreateChangeRequest(b *testing.B) {
	requester := "aura1requester"
	targetDID := "did:aura:target"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		requestID := fmt.Sprintf("%s-%s-%d", requester, targetDID, i)
		_ = sha256.Sum256([]byte(requestID))
	}
}

// BenchmarkCreateRole benchmarks role name validation
func BenchmarkCreateRole(b *testing.B) {
	permissions := []string{"permission1", "permission2"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		roleName := fmt.Sprintf("role-%d", i)
		_ = len(roleName) > 0 && len(permissions) > 0
	}
}
