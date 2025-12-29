// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/aequitas/aura/chain/x/walletsecurity/types"
	wsproto "github.com/aequitas/aura/proto/aura/walletsecurity/v1beta1"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	now := blockTimeToGogoTimestamp(ctx)
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
		return fmt.Errorf("failed to get: %w", err)
	}

	var config wsproto.HardwareWalletConfig
	if err := k.cdc.Unmarshal(configBytes, &config); err != nil {
		return fmt.Errorf("failed to unmarshal hardware wallet config: %w", err)
	}

	config.LastUsedAt = blockTimeToGogoTimestamp(ctx)
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
		return fmt.Errorf("failed to get for ValidateHardwareWalletTransaction: %w", err)
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
	pubKey, sig, err := parseSecpSignature(signature)
	if err != nil {
		return fmt.Errorf("error in validateHardwareWalletSignature for validateColdCardSignature: %w", err)
	}

	digest := sha256.Sum256([]byte("aura-hww:" + address + ":" + deviceID))
	if !pubKey.VerifySignature(digest[:], sig) {
		return types.ErrInvalidDeviceSignature
	}

	if err := ensureAddressMatchesPubKey(address, pubKey); err != nil {
		return fmt.Errorf("error in validateHardwareWalletSignature for ErrInvalidDeviceSignature: %w", err)
	}

	k.logger.Info("validated hardware wallet signature",
		"address", address,
		"device_id", deviceID,
	)

	return nil
}

// validateLedgerSignature validates a Ledger device signature
func (k Keeper) validateLedgerSignature(txData, signature []byte, address string) error {
	pubKey, sig, err := parseSecpSignature(signature)
	if err != nil {
		return fmt.Errorf("error in validateLedgerSignature for device_id: %w", err)
	}

	txDigest := sha256.Sum256(txData)
	if !pubKey.VerifySignature(txDigest[:], sig) {
		return types.ErrInvalidDeviceSignature
	}
	if err := ensureAddressMatchesPubKey(address, pubKey); err != nil {
		return fmt.Errorf("error in validateLedgerSignature for ErrInvalidDeviceSignature: %w", err)
	}

	k.logger.Info("validated Ledger signature",
		"address", address,
		"tx_size", len(txData),
	)

	return nil
}

// validateTrezorSignature validates a Trezor device signature.
// Trezor uses secp256k1 ECDSA signatures with a specific message format.
// The signature format is: 65 bytes (r[32] + s[32] + v[1]) where v is the recovery ID.
func (k Keeper) validateTrezorSignature(txData, signature []byte, address string) error {
	// Trezor signature format: 65 bytes (r[32] || s[32] || v[1])
	// or 64 bytes without recovery ID for some operations
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	// For Trezor, signatures can come as:
	// - 64 bytes: r[32] || s[32] (compact format)
	// - 65 bytes: r[32] || s[32] || v[1] (recoverable format)
	// - 97 bytes: pubkey[33] || sig[64] (with embedded pubkey)

	var pubKey *secp256k1.PubKey
	var sig []byte

	switch len(signature) {
	case 97:
		// Full format: compressed pubkey (33 bytes) + signature (64 bytes)
		pubKey = &secp256k1.PubKey{Key: signature[:33]}
		sig = signature[33:]
	case 65:
		// Recoverable signature: r[32] || s[32] || v[1]
		// We need to recover the public key from the signature and message
		recoveredPubKey, err := k.recoverPubKeyFromSignature(txData, signature)
		if err != nil {
			return fmt.Errorf("failed to recover public key: %w", err)
		}
		pubKey = recoveredPubKey
		sig = signature[:64]
	case 64:
		// Compact format without recovery - need pubkey from address lookup
		// This requires additional context, fail for now
		return fmt.Errorf("compact signature requires pubkey context")
	default:
		return types.ErrInvalidDeviceSignature
	}

	// Compute transaction digest (Trezor uses SHA256 hash of tx data)
	txDigest := sha256.Sum256(txData)

	// Verify the signature
	if !pubKey.VerifySignature(txDigest[:], sig) {
		return types.ErrInvalidDeviceSignature
	}

	// Verify the recovered/provided pubkey matches the registered address
	if err := ensureAddressMatchesPubKey(address, pubKey); err != nil {
		return fmt.Errorf("address mismatch: %w", err)
	}

	k.logger.Info("validated Trezor signature",
		"address", address,
		"tx_size", len(txData),
		"sig_format", len(signature),
	)

	return nil
}

// validateKeepKeySignature validates a KeepKey device signature.
// KeepKey uses the same signature format as Trezor (both are BIP32/44 compatible).
func (k Keeper) validateKeepKeySignature(txData, signature []byte, address string) error {
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	var pubKey *secp256k1.PubKey
	var sig []byte

	switch len(signature) {
	case 97:
		// Full format: compressed pubkey (33 bytes) + signature (64 bytes)
		pubKey = &secp256k1.PubKey{Key: signature[:33]}
		sig = signature[33:]
	case 65:
		// Recoverable signature: r[32] || s[32] || v[1]
		recoveredPubKey, err := k.recoverPubKeyFromSignature(txData, signature)
		if err != nil {
			return fmt.Errorf("failed to recover public key: %w", err)
		}
		pubKey = recoveredPubKey
		sig = signature[:64]
	case 64:
		return fmt.Errorf("compact signature requires pubkey context")
	default:
		return types.ErrInvalidDeviceSignature
	}

	// Compute transaction digest
	txDigest := sha256.Sum256(txData)

	// Verify the signature
	if !pubKey.VerifySignature(txDigest[:], sig) {
		return types.ErrInvalidDeviceSignature
	}

	// Verify address matches
	if err := ensureAddressMatchesPubKey(address, pubKey); err != nil {
		return fmt.Errorf("address mismatch: %w", err)
	}

	k.logger.Info("validated KeepKey signature",
		"address", address,
		"tx_size", len(txData),
		"sig_format", len(signature),
	)

	return nil
}

// validateColdCardSignature validates a ColdCard device signature.
// ColdCard is a Bitcoin-focused hardware wallet using standard secp256k1 signatures.
// For Cosmos compatibility, it uses the same ECDSA verification.
func (k Keeper) validateColdCardSignature(txData, signature []byte, address string) error {
	if len(signature) < 64 {
		return types.ErrInvalidDeviceSignature
	}

	var pubKey *secp256k1.PubKey
	var sig []byte

	switch len(signature) {
	case 97:
		// Full format: compressed pubkey (33 bytes) + signature (64 bytes)
		pubKey = &secp256k1.PubKey{Key: signature[:33]}
		sig = signature[33:]
	case 65:
		// Recoverable signature with recovery ID
		recoveredPubKey, err := k.recoverPubKeyFromSignature(txData, signature)
		if err != nil {
			return fmt.Errorf("failed to recover public key: %w", err)
		}
		pubKey = recoveredPubKey
		sig = signature[:64]
	case 64:
		return fmt.Errorf("compact signature requires pubkey context")
	default:
		return types.ErrInvalidDeviceSignature
	}

	// Compute transaction digest (SHA256 for ColdCard)
	txDigest := sha256.Sum256(txData)

	// Verify the signature
	if !pubKey.VerifySignature(txDigest[:], sig) {
		return types.ErrInvalidDeviceSignature
	}

	// Verify address matches
	if err := ensureAddressMatchesPubKey(address, pubKey); err != nil {
		return fmt.Errorf("address mismatch: %w", err)
	}

	k.logger.Info("validated ColdCard signature",
		"address", address,
		"tx_size", len(txData),
		"sig_format", len(signature),
	)

	return nil
}

// recoverPubKeyFromSignature recovers a secp256k1 public key from a recoverable signature.
// The signature must be 65 bytes: r[32] || s[32] || v[1] where v is the recovery ID (0-3).
func (k Keeper) recoverPubKeyFromSignature(message, signature []byte) (*secp256k1.PubKey, error) {
	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: expected 65, got %d", len(signature))
	}

	// Extract recovery ID from last byte
	recoveryID := signature[64]
	if recoveryID >= 4 {
		// Handle EIP-155 style recovery IDs (27, 28, etc.)
		if recoveryID >= 27 && recoveryID <= 30 {
			recoveryID -= 27
		} else {
			return nil, fmt.Errorf("invalid recovery ID: %d", recoveryID)
		}
	}

	// Compute message hash
	msgHash := sha256.Sum256(message)

	// Use secp256k1 ECDSA recovery
	// The Cosmos SDK secp256k1 package doesn't expose recovery directly,
	// so we implement it using the underlying crypto operations
	pubKeyBytes, err := recoverSecp256k1PubKey(msgHash[:], signature[:64], int(recoveryID))
	if err != nil {
		return nil, fmt.Errorf("public key recovery failed: %w", err)
	}

	return &secp256k1.PubKey{Key: pubKeyBytes}, nil
}

// recoverSecp256k1PubKey recovers the public key from an ECDSA signature.
// This implements SEC1 public key recovery for the secp256k1 curve.
func recoverSecp256k1PubKey(hash, sig []byte, recoveryID int) ([]byte, error) {
	if len(hash) != 32 || len(sig) != 64 {
		return nil, fmt.Errorf("invalid input lengths")
	}

	// Parse r and s from signature
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])

	// secp256k1 curve parameters
	curve := secp256k1Curve()
	n := curve.Params().N

	// Validate r and s are in valid range
	if r.Sign() <= 0 || r.Cmp(n) >= 0 {
		return nil, fmt.Errorf("r is out of range")
	}
	if s.Sign() <= 0 || s.Cmp(n) >= 0 {
		return nil, fmt.Errorf("s is out of range")
	}

	// Calculate the recovery point R
	// R.x = r + (recoveryID >> 1) * n (for recoveryID 0,1 this is just r)
	x := new(big.Int).Set(r)
	if recoveryID >= 2 {
		x.Add(x, n)
	}

	// Calculate y from x using curve equation: y^2 = x^3 + 7 (mod p)
	y := computeSecp256k1Y(x, recoveryID&1 == 1)
	if y == nil {
		return nil, fmt.Errorf("failed to compute y coordinate")
	}

	// Verify point is on curve
	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("recovered point is not on curve")
	}

	// Calculate public key: Q = r^-1 * (s*R - e*G)
	e := new(big.Int).SetBytes(hash)
	rInv := new(big.Int).ModInverse(r, n)
	if rInv == nil {
		return nil, fmt.Errorf("r has no modular inverse")
	}

	// s * R
	sRx, sRy := curve.ScalarMult(x, y, s.Bytes())

	// e * G
	eGx, eGy := curve.ScalarBaseMult(e.Bytes())

	// s*R - e*G
	negEGy := new(big.Int).Sub(curve.Params().P, eGy)
	diffX, diffY := curve.Add(sRx, sRy, eGx, negEGy)

	// r^-1 * (s*R - e*G)
	qX, qY := curve.ScalarMult(diffX, diffY, rInv.Bytes())

	// Marshal to compressed format (33 bytes)
	pubKey := make([]byte, 33)
	if qY.Bit(0) == 0 {
		pubKey[0] = 0x02
	} else {
		pubKey[0] = 0x03
	}
	qXBytes := qX.Bytes()
	copy(pubKey[33-len(qXBytes):], qXBytes)

	return pubKey, nil
}

// secp256k1Curve returns the secp256k1 curve parameters.
// This is a helper since Go's crypto/elliptic doesn't include secp256k1.
func secp256k1Curve() *secp256k1CurveParams {
	return &secp256k1Params
}

type secp256k1CurveParams struct {
	P       *big.Int
	N       *big.Int
	B       *big.Int
	Gx, Gy  *big.Int
	BitSize int
}

var secp256k1Params = secp256k1CurveParams{
	P:       fromHex("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f"),
	N:       fromHex("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141"),
	B:       big.NewInt(7),
	Gx:      fromHex("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"),
	Gy:      fromHex("483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"),
	BitSize: 256,
}

func (curve *secp256k1CurveParams) Params() *secp256k1CurveParams {
	return curve
}

func (curve *secp256k1CurveParams) IsOnCurve(x, y *big.Int) bool {
	// y^2 = x^3 + 7 (mod p)
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, curve.P)

	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)

	return y2.Cmp(x3) == 0
}

func (curve *secp256k1CurveParams) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	// Point addition on secp256k1
	if x1.Sign() == 0 && y1.Sign() == 0 {
		return x2, y2
	}
	if x2.Sign() == 0 && y2.Sign() == 0 {
		return x1, y1
	}

	p := curve.P
	lambda := new(big.Int)

	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		// Point doubling: lambda = (3*x1^2) / (2*y1)
		x1Sq := new(big.Int).Mul(x1, x1)
		x1Sq.Mul(x1Sq, big.NewInt(3))
		x1Sq.Mod(x1Sq, p)

		twoY1 := new(big.Int).Mul(y1, big.NewInt(2))
		twoY1Inv := new(big.Int).ModInverse(twoY1, p)
		if twoY1Inv == nil {
			return big.NewInt(0), big.NewInt(0)
		}
		lambda.Mul(x1Sq, twoY1Inv)
		lambda.Mod(lambda, p)
	} else {
		// Point addition: lambda = (y2-y1) / (x2-x1)
		dy := new(big.Int).Sub(y2, y1)
		dx := new(big.Int).Sub(x2, x1)
		dxInv := new(big.Int).ModInverse(dx, p)
		if dxInv == nil {
			return big.NewInt(0), big.NewInt(0)
		}
		lambda.Mul(dy, dxInv)
		lambda.Mod(lambda, p)
	}

	// x3 = lambda^2 - x1 - x2
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)

	// y3 = lambda*(x1-x3) - y1
	y3 := new(big.Int).Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}

func (curve *secp256k1CurveParams) ScalarMult(x1, y1 *big.Int, k []byte) (*big.Int, *big.Int) {
	// Double-and-add scalar multiplication
	kInt := new(big.Int).SetBytes(k)
	kInt.Mod(kInt, curve.N)

	rx, ry := big.NewInt(0), big.NewInt(0)
	tx, ty := new(big.Int).Set(x1), new(big.Int).Set(y1)

	for i := 0; i < kInt.BitLen(); i++ {
		if kInt.Bit(i) == 1 {
			rx, ry = curve.Add(rx, ry, tx, ty)
		}
		tx, ty = curve.Add(tx, ty, tx, ty)
	}

	return rx, ry
}

func (curve *secp256k1CurveParams) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return curve.ScalarMult(curve.Gx, curve.Gy, k)
}

// computeSecp256k1Y computes the y-coordinate for a given x on secp256k1.
func computeSecp256k1Y(x *big.Int, odd bool) *big.Int {
	p := secp256k1Params.P

	// y^2 = x^3 + 7 (mod p)
	y2 := new(big.Int).Mul(x, x)
	y2.Mul(y2, x)
	y2.Add(y2, big.NewInt(7))
	y2.Mod(y2, p)

	// Compute modular square root using Tonelli-Shanks
	// For secp256k1, p ≡ 3 (mod 4), so y = y2^((p+1)/4) mod p
	exp := new(big.Int).Add(p, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	y := new(big.Int).Exp(y2, exp, p)

	// Verify
	check := new(big.Int).Mul(y, y)
	check.Mod(check, p)
	if check.Cmp(y2) != 0 {
		return nil // No valid y exists
	}

	// Adjust for odd/even
	if (y.Bit(0) == 1) != odd {
		y.Sub(p, y)
	}

	return y
}

func fromHex(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 16)
	return n
}

// parseSecpSignature expects a compressed secp256k1 pubkey (33 bytes) followed by a 64-byte signature.
func parseSecpSignature(signature []byte) (*secp256k1.PubKey, []byte, error) {
	if len(signature) != 97 {
		return nil, nil, types.ErrInvalidDeviceSignature
	}
	pubKeyBytes := signature[:33]
	sig := signature[33:]
	if len(sig) != 64 {
		return nil, nil, types.ErrInvalidDeviceSignature
	}

	pubKey := &secp256k1.PubKey{Key: pubKeyBytes}
	return pubKey, sig, nil
}

// ensureAddressMatchesPubKey confirms the bech32 account address matches the provided pubkey.
func ensureAddressMatchesPubKey(address string, pubKey *secp256k1.PubKey) error {
	accAddr := sdk.AccAddress(pubKey.Address())
	if accAddr.String() != address {
		return types.ErrInvalidDeviceSignature
	}
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
