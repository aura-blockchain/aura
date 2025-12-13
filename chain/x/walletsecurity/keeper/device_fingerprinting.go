package keeper

import (
	"context"
	"crypto/sha256"
	"fmt"

	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"
)

// RegisterDevice registers a new device for a wallet
func (k Keeper) RegisterDevice(ctx context.Context, walletID, deviceID, deviceName, deviceType string, fingerprint []byte) (*wsproto.DeviceFingerprint, error) {
	device := &wsproto.DeviceFingerprint{
		DeviceId:        deviceID,
		WalletId:        walletID,
		DeviceName:      deviceName,
		DeviceType:      deviceType,
		Fingerprint:     fingerprint,
		FingerprintHash: k.hashFingerprint(fingerprint),
		Trusted:         true,
		RegisteredAt:    blockTimeToGogoTimestamp(ctx),
		LastSeenAt:      blockTimeToGogoTimestamp(ctx),
	}

	deviceBytes, err := k.cdc.Marshal(device)
	if err != nil {
		return nil, err
	}

	store := k.getStore(ctx)
	key := []byte(fmt.Sprintf("device_%s_%s", walletID, deviceID))
	if err := store.Set(key, deviceBytes); err != nil {
		return nil, err
	}

	return device, nil
}

// VerifyDevice verifies a device fingerprint
func (k Keeper) VerifyDevice(ctx context.Context, walletID, deviceID string, fingerprint []byte) (bool, error) {
	store := k.getStore(ctx)
	key := []byte(fmt.Sprintf("device_%s_%s", walletID, deviceID))

	deviceBytes, err := store.Get(key)
	if err != nil {
		return false, err
	}
	if deviceBytes == nil {
		return false, fmt.Errorf("device not registered")
	}

	var device wsproto.DeviceFingerprint
	if err := k.cdc.Unmarshal(deviceBytes, &device); err != nil {
		return false, err
	}

	if !device.Trusted {
		return false, fmt.Errorf("device not trusted")
	}

	// Verify fingerprint hash
	providedHash := k.hashFingerprint(fingerprint)
	if string(providedHash) != string(device.FingerprintHash) {
		return false, fmt.Errorf("fingerprint mismatch")
	}

	// Update last seen
	device.LastSeenAt = blockTimeToGogoTimestamp(ctx)
	updatedBytes, err := k.cdc.Marshal(&device)
	if err != nil {
		return false, err
	}
	if err := store.Set(key, updatedBytes); err != nil {
		return false, err
	}

	return true, nil
}

// RevokeDevice revokes a device's access
func (k Keeper) RevokeDevice(ctx context.Context, walletID, deviceID string) error {
	store := k.getStore(ctx)
	key := []byte(fmt.Sprintf("device_%s_%s", walletID, deviceID))

	deviceBytes, err := store.Get(key)
	if err != nil {
		return err
	}
	if deviceBytes == nil {
		return fmt.Errorf("device not found")
	}

	var device wsproto.DeviceFingerprint
	if err := k.cdc.Unmarshal(deviceBytes, &device); err != nil {
		return err
	}

	device.Trusted = false
	device.RevokedAt = blockTimeToGogoTimestamp(ctx)

	updatedBytes, err := k.cdc.Marshal(&device)
	if err != nil {
		return err
	}

	return store.Set(key, updatedBytes)
}

func (k Keeper) hashFingerprint(fingerprint []byte) []byte {
	hash := sha256.Sum256(fingerprint)
	return hash[:]
}

// GetDevices retrieves all devices for a wallet
func (k Keeper) GetDevices(ctx context.Context, walletID string) ([]*wsproto.DeviceFingerprint, error) {
	kvStore := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("device_%s_", walletID))

	// Create iterator manually
	var devices []*wsproto.DeviceFingerprint
	iter, err := kvStore.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Check if key starts with prefix
		if !hasPrefix(iter.Key(), prefix) {
			break
		}

		var device wsproto.DeviceFingerprint
		if err := k.cdc.Unmarshal(iter.Value(), &device); err != nil {
			continue
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

// DetectAnomalousDevice detects if a device shows anomalous behavior
func (k Keeper) DetectAnomalousDevice(ctx context.Context, walletID, deviceID string) (bool, error) {
	devices, err := k.GetDevices(ctx, walletID)
	if err != nil {
		return false, err
	}

	var targetDevice *wsproto.DeviceFingerprint
	for _, d := range devices {
		if d.DeviceId == deviceID {
			targetDevice = d
			break
		}
	}

	if targetDevice == nil {
		return true, nil // Unknown device is anomalous
	}

	// Check for rapid location changes (if geolocation data available)
	// Check for unusual access patterns
	// For now, simple check: if device was revoked
	return !targetDevice.Trusted, nil
}
