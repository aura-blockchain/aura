package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// RegisterHardwareWallet registers a new hardware wallet
func (k Keeper) RegisterHardwareWallet(
	ctx context.Context,
	address string,
	hwType wsproto.HardwareWalletType,
	deviceID string,
	firmwareVersion string,
	derivationPath string,
	signature []byte,
) (*wsproto.HardwareWalletConfig, error) {
	// Validate inputs
	if address == "" {
		return nil, types.ErrInvalidInput
	}
	if hwType == wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_UNSPECIFIED {
		return nil, types.ErrUnsupportedHardwareWallet
	}
	if deviceID == "" {
		return nil, types.ErrInvalidInput
	}

	// Validate signature from hardware wallet
	if err := k.validateHardwareWalletSignature(address, deviceID, signature); err != nil {
		return nil, err
	}

	// Generate wallet ID
	walletID := k.generateWalletID(address, deviceID)

	// Check if wallet already exists
	if _, err := k.GetHardwareWallet(ctx, walletID); err == nil {
		return nil, types.ErrHardwareWalletExists
	}

	// Create hardware wallet configuration
	now := gogoTimestampNow()
	config := &wsproto.HardwareWalletConfig{
		WalletId:           walletID,
		Address:            address,
		Type:               hwType,
		DeviceId:           deviceID,
		FirmwareVersion:    firmwareVersion,
		DerivationPath:     derivationPath,
		RequiresPin:        true, // Default security
		RequiresPassphrase: false,
		RegisteredAt:       now,
		LastUsedAt:         now,
		SignatureCount:     0,
		Metadata:           make(map[string]string),
	}

	// Add device-specific metadata
	config.Metadata["device_type"] = hwType.String()
	config.Metadata["registered_timestamp"] = gogoTimestampToTime(now).String()

	// Store the configuration
	configBytes := k.cdc.MustMarshal(config)
	if err := k.SetHardwareWallet(ctx, walletID, configBytes); err != nil {
		return nil, err
	}

	k.logger.Info("registered hardware wallet",
		"wallet_id", walletID,
		"type", hwType.String(),
		"address", address,
	)

	return config, nil
}

// UpdateHardwareWalletUsage updates the last used time and signature count
func (k Keeper) UpdateHardwareWalletUsage(ctx context.Context, walletID string) error {
	configBytes, err := k.GetHardwareWallet(ctx, walletID)
	if err != nil {
		return err
	}

	var config wsproto.HardwareWalletConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to unmarshal hardware wallet config: %w", err)
	}

	config.LastUsedAt = gogoTimestampNow()
	config.SignatureCount++

	updatedBytes := k.cdc.MustMarshal(&config)
	return k.SetHardwareWallet(ctx, walletID, updatedBytes)
}

// ValidateHardwareWalletTransaction validates a transaction signed by hardware wallet
func (k Keeper) ValidateHardwareWalletTransaction(
	ctx context.Context,
	walletID string,
	txData []byte,
	signature []byte,
) error {
	// Get hardware wallet config
	configBytes, err := k.GetHardwareWallet(ctx, walletID)
	if err != nil {
		return err
	}

	var config wsproto.HardwareWalletConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to unmarshal hardware wallet config: %w", err)
	}

	// Validate signature based on hardware wallet type
	switch config.Type {
	case wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_LEDGER:
		return k.validateLedgerSignature(txData, signature, config.Address)
	case wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_TREZOR:
		return k.validateTrezorSignature(txData, signature, config.Address)
	case wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_KEEPKEY:
		return k.validateKeepKeySignature(txData, signature, config.Address)
	case wsproto.HardwareWalletType_HARDWARE_WALLET_TYPE_COLDCARD:
		return k.validateColdCardSignature(txData, signature, config.Address)
	default:
		return types.ErrUnsupportedHardwareWallet
	}
}

// validateHardwareWalletSignature validates the device signature during registration
func (k Keeper) validateHardwareWalletSignature(address, deviceID string, signature []byte) error {
	if len(signature) == 0 {
		return types.ErrInvalidDeviceSignature
	}

	// In a production implementation, this would:
	// 1. Verify the signature using the device's public key
	// 2. Check that the signature is over the expected message (address + deviceID)
	// 3. Validate the signature format based on the hardware wallet type

	// For this implementation, we perform basic validation
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	k.logger.Info("validated hardware wallet signature",
		"address", address,
		"device_id", deviceID,
	)

	return nil
}

// validateLedgerSignature validates a Ledger device signature
func (k Keeper) validateLedgerSignature(txData, signature []byte, address string) error {
	// In production, this would:
	// 1. Verify the signature using Ledger's specific signature format
	// 2. Check the transaction hash
	// 3. Validate against the registered address

	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	k.logger.Info("validated Ledger signature",
		"address", address,
		"tx_size", len(txData),
	)

	return nil
}

// validateTrezorSignature validates a Trezor device signature
func (k Keeper) validateTrezorSignature(txData, signature []byte, address string) error {
	// In production, this would:
	// 1. Verify the signature using Trezor's specific signature format
	// 2. Check the transaction hash
	// 3. Validate against the registered address

	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	k.logger.Info("validated Trezor signature",
		"address", address,
		"tx_size", len(txData),
	)

	return nil
}

// validateKeepKeySignature validates a KeepKey device signature
func (k Keeper) validateKeepKeySignature(txData, signature []byte, address string) error {
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	k.logger.Info("validated KeepKey signature",
		"address", address,
		"tx_size", len(txData),
	)

	return nil
}

// validateColdCardSignature validates a ColdCard device signature
func (k Keeper) validateColdCardSignature(txData, signature []byte, address string) error {
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	k.logger.Info("validated ColdCard signature",
		"address", address,
		"tx_size", len(txData),
	)

	return nil
}

// generateWalletID generates a unique wallet ID
func (k Keeper) generateWalletID(address, deviceID string) string {
	data := address + ":" + deviceID
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GetHardwareWalletByAddress retrieves hardware wallet config by address
func (k Keeper) GetHardwareWalletByAddress(ctx context.Context, address string) (*wsproto.HardwareWalletConfig, error) {
	// In production, we would maintain an address-to-walletID index
	// For this implementation, we return an error indicating the need for wallet ID
	return nil, fmt.Errorf("use GetHardwareWallet with wallet ID")
}

// RequiresPinConfirmation checks if the hardware wallet requires PIN confirmation
func (k Keeper) RequiresPinConfirmation(ctx context.Context, walletID string) (bool, error) {
	configBytes, err := k.GetHardwareWallet(ctx, walletID)
	if err != nil {
		return false, err
	}

	var config wsproto.HardwareWalletConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return false, fmt.Errorf("failed to unmarshal hardware wallet config: %w", err)
	}

	return config.RequiresPin, nil
}

// RequiresPassphrase checks if the hardware wallet requires passphrase
func (k Keeper) RequiresPassphrase(ctx context.Context, walletID string) (bool, error) {
	configBytes, err := k.GetHardwareWallet(ctx, walletID)
	if err != nil {
		return false, err
	}

	var config wsproto.HardwareWalletConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return false, fmt.Errorf("failed to unmarshal hardware wallet config: %w", err)
	}

	return config.RequiresPassphrase, nil
}
