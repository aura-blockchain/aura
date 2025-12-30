// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "walletsecurity", ModuleName)
	require.Equal(t, ModuleName, StoreKey)
	require.Equal(t, ModuleName, RouterKey)
	require.Equal(t, ModuleName, QuerierRoute)
}

func TestKeyPrefixes_Unique(t *testing.T) {
	prefixes := [][]byte{
		HardwareWalletPrefix,
		MultiSigWalletPrefix,
		PendingMultiSigTxPrefix,
		SocialRecoveryPrefix,
		RecoveryRequestPrefix,
		SpendingLimitPrefix,
		SessionConfigPrefix,
		BiometricAuthPrefix,
		SecureEnclavePrefix,
		EncryptedBackupPrefix,
		DustFilterPrefix,
		DomainVerificationPrefix,
		SecurityMetricsPrefix,
		DustTransactionPrefix,
	}

	// Ensure all prefixes are unique
	for i, p1 := range prefixes {
		for j, p2 := range prefixes {
			if i != j {
				require.NotEqual(t, p1[0], p2[0], "prefixes at index %d and %d should have different values", i, j)
			}
		}
	}
}

func TestKeyPrefixes_Values(t *testing.T) {
	require.Equal(t, byte(0x01), HardwareWalletPrefix[0])
	require.Equal(t, byte(0x02), MultiSigWalletPrefix[0])
	require.Equal(t, byte(0x03), PendingMultiSigTxPrefix[0])
	require.Equal(t, byte(0x04), SocialRecoveryPrefix[0])
	require.Equal(t, byte(0x05), RecoveryRequestPrefix[0])
	require.Equal(t, byte(0x06), SpendingLimitPrefix[0])
	require.Equal(t, byte(0x07), SessionPrefix[0])
	require.Equal(t, byte(0x14), SessionConfigPrefix[0]) // Unique value to avoid collision with SessionPrefix
	require.Equal(t, byte(0x08), BiometricAuthPrefix[0])
	require.Equal(t, byte(0x09), SecureEnclavePrefix[0])
	require.Equal(t, byte(0x0a), EncryptedBackupPrefix[0])
	require.Equal(t, byte(0x0b), DustFilterPrefix[0])
	require.Equal(t, byte(0x0c), DomainVerificationPrefix[0])
	require.Equal(t, byte(0x0d), SecurityMetricsPrefix[0])
	require.Equal(t, byte(0x0e), DustTransactionPrefix[0])
}

func TestGetHardwareWalletKey(t *testing.T) {
	walletID := "wallet123"
	key := GetHardwareWalletKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, HardwareWalletPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetMultiSigWalletKey(t *testing.T) {
	walletID := "multisig456"
	key := GetMultiSigWalletKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, MultiSigWalletPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetPendingMultiSigTxKey(t *testing.T) {
	txID := "tx789"
	key := GetPendingMultiSigTxKey(txID)

	require.NotEmpty(t, key)
	require.Equal(t, PendingMultiSigTxPrefix[0], key[0])
	require.Contains(t, string(key), txID)
}

func TestGetSocialRecoveryKey(t *testing.T) {
	walletID := "recovery123"
	key := GetSocialRecoveryKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, SocialRecoveryPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetRecoveryRequestKey(t *testing.T) {
	requestID := "request456"
	key := GetRecoveryRequestKey(requestID)

	require.NotEmpty(t, key)
	require.Equal(t, RecoveryRequestPrefix[0], key[0])
	require.Contains(t, string(key), requestID)
}

func TestGetSpendingLimitKey(t *testing.T) {
	walletID := "wallet123"
	denom := "uaura"
	key := GetSpendingLimitKey(walletID, denom)

	require.NotEmpty(t, key)
	require.Equal(t, SpendingLimitPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
	require.Contains(t, string(key), denom)
	require.Contains(t, string(key), ":")
}

func TestGetSessionConfigKey(t *testing.T) {
	sessionID := "session789"
	key := GetSessionConfigKey(sessionID)

	require.NotEmpty(t, key)
	require.Equal(t, SessionConfigPrefix[0], key[0])
	require.Contains(t, string(key), sessionID)
}

func TestGetBiometricAuthKey(t *testing.T) {
	walletID := "biometric123"
	key := GetBiometricAuthKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, BiometricAuthPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetSecureEnclaveKey(t *testing.T) {
	walletID := "enclave456"
	key := GetSecureEnclaveKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, SecureEnclavePrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetEncryptedBackupKey(t *testing.T) {
	backupID := "backup789"
	key := GetEncryptedBackupKey(backupID)

	require.NotEmpty(t, key)
	require.Equal(t, EncryptedBackupPrefix[0], key[0])
	require.Contains(t, string(key), backupID)
}

func TestGetDustFilterKey(t *testing.T) {
	walletID := "dust123"
	key := GetDustFilterKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, DustFilterPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetDomainVerificationKey(t *testing.T) {
	domain := "example.com"
	key := GetDomainVerificationKey(domain)

	require.NotEmpty(t, key)
	require.Equal(t, DomainVerificationPrefix[0], key[0])
	require.Contains(t, string(key), domain)
}

func TestGetSecurityMetricsKey(t *testing.T) {
	walletID := "metrics456"
	key := GetSecurityMetricsKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, SecurityMetricsPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetDustTransactionKey(t *testing.T) {
	txHash := "0xabc123"
	key := GetDustTransactionKey(txHash)

	require.NotEmpty(t, key)
	require.Equal(t, DustTransactionPrefix[0], key[0])
	require.Contains(t, string(key), txHash)
}

func TestKeyFunctions_EmptyInput(t *testing.T) {
	// Test that key functions handle empty strings
	tests := []struct {
		name string
		fn   func(string) []byte
	}{
		{"GetHardwareWalletKey", GetHardwareWalletKey},
		{"GetMultiSigWalletKey", GetMultiSigWalletKey},
		{"GetPendingMultiSigTxKey", GetPendingMultiSigTxKey},
		{"GetSocialRecoveryKey", GetSocialRecoveryKey},
		{"GetRecoveryRequestKey", GetRecoveryRequestKey},
		{"GetSessionConfigKey", GetSessionConfigKey},
		{"GetBiometricAuthKey", GetBiometricAuthKey},
		{"GetSecureEnclaveKey", GetSecureEnclaveKey},
		{"GetEncryptedBackupKey", GetEncryptedBackupKey},
		{"GetDustFilterKey", GetDustFilterKey},
		{"GetDomainVerificationKey", GetDomainVerificationKey},
		{"GetSecurityMetricsKey", GetSecurityMetricsKey},
		{"GetDustTransactionKey", GetDustTransactionKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := tt.fn("")
			require.NotEmpty(t, key)
			require.Len(t, key, 1) // Should only contain the prefix
		})
	}
}

func TestGetSpendingLimitKey_EmptyInputs(t *testing.T) {
	tests := []struct {
		name     string
		walletID string
		denom    string
	}{
		{"empty wallet", "", "uaura"},
		{"empty denom", "wallet123", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GetSpendingLimitKey(tt.walletID, tt.denom)
			require.NotEmpty(t, key)
			require.Contains(t, string(key), ":")
		})
	}
}

func TestKeyFunctions_UniqueKeys(t *testing.T) {
	// Test that different IDs produce different keys
	id1 := "id1"
	id2 := "id2"

	key1 := GetHardwareWalletKey(id1)
	key2 := GetHardwareWalletKey(id2)
	require.NotEqual(t, key1, key2)

	key1 = GetMultiSigWalletKey(id1)
	key2 = GetMultiSigWalletKey(id2)
	require.NotEqual(t, key1, key2)
}

func TestGetSessionKey(t *testing.T) {
	sessionID := "session123"
	key := GetSessionKey(sessionID)

	require.NotEmpty(t, key)
	require.Equal(t, SessionPrefix[0], key[0])
	require.Contains(t, string(key), sessionID)
}

func TestGetSessionKey_EmptyInput(t *testing.T) {
	key := GetSessionKey("")
	require.NotEmpty(t, key)
	require.Len(t, key, 1) // Only prefix
}

func TestGetDeviceFingerprintKey(t *testing.T) {
	deviceID := "device456"
	key := GetDeviceFingerprintKey(deviceID)

	require.NotEmpty(t, key)
	require.Equal(t, DeviceFingerprintPrefix[0], key[0])
	require.Contains(t, string(key), deviceID)
}

func TestGetDeviceFingerprintKey_EmptyInput(t *testing.T) {
	key := GetDeviceFingerprintKey("")
	require.NotEmpty(t, key)
	require.Len(t, key, 1)
}

func TestGetAnomalyKey(t *testing.T) {
	anomalyID := "anomaly789"
	key := GetAnomalyKey(anomalyID)

	require.NotEmpty(t, key)
	require.Equal(t, AnomalyPrefix[0], key[0])
	require.Contains(t, string(key), anomalyID)
}

func TestGetAnomalyKey_EmptyInput(t *testing.T) {
	key := GetAnomalyKey("")
	require.NotEmpty(t, key)
	require.Len(t, key, 1)
}

func TestGetWalletAnalyticsKey(t *testing.T) {
	walletID := "analytics123"
	key := GetWalletAnalyticsKey(walletID)

	require.NotEmpty(t, key)
	require.Equal(t, WalletAnalyticsPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
}

func TestGetWalletAnalyticsKey_EmptyInput(t *testing.T) {
	key := GetWalletAnalyticsKey("")
	require.NotEmpty(t, key)
	require.Len(t, key, 1)
}

func TestGetInsurancePolicyKey(t *testing.T) {
	policyID := "policy456"
	key := GetInsurancePolicyKey(policyID)

	require.NotEmpty(t, key)
	require.Equal(t, InsurancePolicyPrefix[0], key[0])
	require.Contains(t, string(key), policyID)
}

func TestGetInsurancePolicyKey_EmptyInput(t *testing.T) {
	key := GetInsurancePolicyKey("")
	require.NotEmpty(t, key)
	require.Len(t, key, 1)
}

func TestGetBiometricProofKey(t *testing.T) {
	walletID := "wallet789"
	proofHash := []byte("proof123hash")
	key := GetBiometricProofKey(walletID, proofHash)

	require.NotEmpty(t, key)
	require.Equal(t, BiometricProofPrefix[0], key[0])
	require.Contains(t, string(key), walletID)
	require.Contains(t, string(key), ":")
	require.Contains(t, string(key), string(proofHash))
}

func TestGetBiometricProofKey_EmptyInputs(t *testing.T) {
	tests := []struct {
		name      string
		walletID  string
		proofHash []byte
	}{
		{"empty wallet", "", []byte("proof")},
		{"empty proof", "wallet", nil},
		{"both empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GetBiometricProofKey(tt.walletID, tt.proofHash)
			require.NotEmpty(t, key)
			require.Contains(t, string(key), ":")
		})
	}
}

func TestGetBiometricProofKey_DifferentProofs(t *testing.T) {
	walletID := "wallet1"
	proof1 := []byte("proof1")
	proof2 := []byte("proof2")

	key1 := GetBiometricProofKey(walletID, proof1)
	key2 := GetBiometricProofKey(walletID, proof2)

	require.NotEqual(t, key1, key2)
}

func TestGetAuthRateLimitKey(t *testing.T) {
	blockHeight := int64(12345)
	address := "aura1test123"
	key := GetAuthRateLimitKey(blockHeight, address)

	require.NotEmpty(t, key)
	require.Contains(t, string(key), AuthRateLimitPrefix)
	require.Contains(t, string(key), "12345")
	require.Contains(t, string(key), address)
}

func TestGetAuthRateLimitKey_DifferentHeights(t *testing.T) {
	address := "aura1test"

	key1 := GetAuthRateLimitKey(100, address)
	key2 := GetAuthRateLimitKey(200, address)

	require.NotEqual(t, key1, key2)
}

func TestGetAuthRateLimitKey_DifferentAddresses(t *testing.T) {
	blockHeight := int64(100)

	key1 := GetAuthRateLimitKey(blockHeight, "addr1")
	key2 := GetAuthRateLimitKey(blockHeight, "addr2")

	require.NotEqual(t, key1, key2)
}
