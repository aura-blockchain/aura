package keeper

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aequitas/aura/chain/x/cryptography/types"
	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// AddCertificatePin adds a certificate pin for a hostname
func (k Keeper) AddCertificatePin(
	ctx context.Context,
	creator string,
	hostname string,
	certificateHashes [][]byte,
	pinType cryptoproto.CertificatePinType,
	expiresAt *time.Time,
) (string, error) {
	if hostname == "" {
		return "", fmt.Errorf("hostname cannot be empty")
	}
	if len(certificateHashes) == 0 {
		return "", fmt.Errorf("at least one certificate hash required")
	}

	// Validate certificate hashes
	for _, hash := range certificateHashes {
		if len(hash) != 32 { // SHA-256 hash
			return "", types.ErrInvalidCertificateHash
		}
	}

	// Generate pin ID
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	pinID := fmt.Sprintf("pin_%s_%d", hostname, blockTime.Unix())

	// Set default expiration if not provided
	var expiresAtProto *timestamppb.Timestamp
	if expiresAt == nil {
		params, _ := k.GetParams(ctx)
		defaultExpiry := blockTime.AddDate(0, 0, int(params.CertificatePinValidityDays))
		expiresAtProto = timestamppb.New(defaultExpiry)
	} else {
		expiresAtProto = timestamppb.New(*expiresAt)
	}

	pin := &cryptoproto.CertificatePin{
		PinId:             pinID,
		Hostname:          hostname,
		CertificateHashes: certificateHashes,
		PinType:           pinType,
		CreatedAt:         timestamppb.New(blockTime),
		ExpiresAt:         expiresAtProto,
		Enabled:           true,
	}

	// Store in KV store
	if err := k.SetCertificatePin(ctx, pin); err != nil {
		return "", err
	}

	k.Logger(ctx).Info("added certificate pin",
		"pin_id", pinID,
		"hostname", hostname,
		"pin_type", pinType.String(),
		"num_hashes", len(certificateHashes),
	)

	return pinID, nil
}

// Note: GetCertificatePin is now implemented in keeper.go using KV store

// VerifyCertificatePin verifies a certificate against pinned hashes
func (k Keeper) VerifyCertificatePin(
	ctx context.Context,
	hostname string,
	certificate []byte,
) (bool, error) {
	pin, err := k.GetCertificatePin(ctx, hostname)
	if err != nil {
		return false, err
	}

	if !pin.Enabled {
		return false, fmt.Errorf("certificate pin disabled for %s", hostname)
	}

	// Check expiration
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	if pin.ExpiresAt != nil && pin.ExpiresAt.AsTime().Before(blockTime) {
		return false, fmt.Errorf("certificate pin expired for %s", hostname)
	}

	// Compute certificate hash based on pin type
	var certHash []byte
	switch pin.PinType {
	case cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_SPKI:
		certHash, err = k.extractSPKIHash(certificate)
	case cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_FULL_CERT:
		certHash = k.hashCertificate(certificate)
	case cryptoproto.CertificatePinType_CERTIFICATE_PIN_TYPE_INTERMEDIATE:
		certHash = k.hashCertificate(certificate)
	default:
		return false, fmt.Errorf("unsupported pin type")
	}

	if err != nil {
		return false, err
	}

	// Check if hash matches any pinned hash
	for _, pinnedHash := range pin.CertificateHashes {
		if k.CompareHashes(certHash, pinnedHash) {
			k.Logger(ctx).Info("certificate pin verified",
				"hostname", hostname,
				"pin_type", pin.PinType.String(),
			)
			return true, nil
		}
	}

	k.Logger(ctx).Warn("certificate pin verification failed",
		"hostname", hostname,
		"pin_type", pin.PinType.String(),
	)

	return false, types.ErrCertificateVerificationFailed
}

// extractSPKIHash extracts and hashes the Subject Public Key Info from a certificate
func (k Keeper) extractSPKIHash(certBytes []byte) ([]byte, error) {
	// Parse certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, err
	}

	// Extract SPKI (SubjectPublicKeyInfo)
	spki := cert.RawSubjectPublicKeyInfo

	// Hash SPKI
	h := sha256.New()
	h.Write(spki)
	return h.Sum(nil), nil
}

// hashCertificate computes SHA-256 hash of a certificate
func (k Keeper) hashCertificate(certBytes []byte) []byte {
	h := sha256.New()
	h.Write(certBytes)
	return h.Sum(nil)
}

// UpdateCertificatePin updates an existing certificate pin
func (k Keeper) UpdateCertificatePin(
	ctx context.Context,
	hostname string,
	certificateHashes [][]byte,
	expiresAt *time.Time,
) error {
	pin, err := k.GetCertificatePin(ctx, hostname)
	if err != nil {
		return err
	}

	// Update hashes if provided
	if len(certificateHashes) > 0 {
		// Validate new hashes
		for _, hash := range certificateHashes {
			if len(hash) != 32 {
				return types.ErrInvalidCertificateHash
			}
		}
		pin.CertificateHashes = certificateHashes
	}

	// Update expiration if provided
	if expiresAt != nil {
		pin.ExpiresAt = timestamppb.New(*expiresAt)
	}

	// Store updated pin
	if err := k.SetCertificatePin(ctx, pin); err != nil {
		return err
	}

	k.Logger(ctx).Info("updated certificate pin",
		"hostname", hostname,
	)

	return nil
}

// RemoveCertificatePin removes a certificate pin for a hostname
func (k Keeper) RemoveCertificatePin(ctx context.Context, hostname string) error {
	if err := k.DeleteCertificatePin(ctx, hostname); err != nil {
		return err
	}

	k.Logger(ctx).Info("removed certificate pin",
		"hostname", hostname,
	)

	return nil
}

// EnableCertificatePin enables a certificate pin
func (k Keeper) EnableCertificatePin(ctx context.Context, hostname string) error {
	pin, err := k.GetCertificatePin(ctx, hostname)
	if err != nil {
		return err
	}

	pin.Enabled = true

	if err := k.SetCertificatePin(ctx, pin); err != nil {
		return err
	}

	k.Logger(ctx).Info("enabled certificate pin", "hostname", hostname)

	return nil
}

// DisableCertificatePin disables a certificate pin
func (k Keeper) DisableCertificatePin(ctx context.Context, hostname string) error {
	pin, err := k.GetCertificatePin(ctx, hostname)
	if err != nil {
		return err
	}

	pin.Enabled = false

	if err := k.SetCertificatePin(ctx, pin); err != nil {
		return err
	}

	k.Logger(ctx).Info("disabled certificate pin", "hostname", hostname)

	return nil
}

// Note: ListCertificatePins is now implemented in keeper.go using KV store

// RotateCertificatePin rotates certificate hashes for a hostname
func (k Keeper) RotateCertificatePin(
	ctx context.Context,
	hostname string,
	newCertificateHashes [][]byte,
) error {
	pin, err := k.GetCertificatePin(ctx, hostname)
	if err != nil {
		return err
	}

	// Validate new hashes
	for _, hash := range newCertificateHashes {
		if len(hash) != 32 {
			return types.ErrInvalidCertificateHash
		}
	}

	// Append new hashes (keeping old ones temporarily for rollout)
	pin.CertificateHashes = append(pin.CertificateHashes, newCertificateHashes...)

	// Store updated pin
	if err := k.SetCertificatePin(ctx, pin); err != nil {
		return err
	}

	k.Logger(ctx).Info("rotated certificate pin",
		"hostname", hostname,
		"total_hashes", len(pin.CertificateHashes),
	)

	return nil
}

// CleanupExpiredPins removes expired certificate pins
func (k Keeper) CleanupExpiredPins(ctx context.Context) error {
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime()
	expired := []string{}

	_ = k.IterateCertificatePins(ctx, func(pin *cryptoproto.CertificatePin) bool {
		if pin.ExpiresAt != nil && pin.ExpiresAt.AsTime().Before(blockTime) {
			expired = append(expired, pin.Hostname)
		}
		return false
	})

	for _, hostname := range expired {
		if err := k.DeleteCertificatePin(ctx, hostname); err != nil {
			return err
		}

		k.Logger(ctx).Info("removed expired certificate pin",
			"hostname", hostname,
		)
	}

	return nil
}

// Note: SetCertificatePin is now implemented in keeper.go using KV store
