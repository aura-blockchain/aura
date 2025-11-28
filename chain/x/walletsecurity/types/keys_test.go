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
	require.Equal(t, byte(0x07), SessionConfigPrefix[0])
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
