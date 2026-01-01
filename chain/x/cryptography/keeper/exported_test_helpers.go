// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"

	cryptoproto "github.com/aequitas/aura/proto/aura/cryptography/v1beta1"
)

// EncryptAESGCMExported exports encryptAESGCM for testing
func (k Keeper) EncryptAESGCMExported(plaintext, key []byte) ([]byte, error) {
	return k.encryptAESGCM(plaintext, key)
}

// ExtractSPKIHashExported exports extractSPKIHash for testing
func (k Keeper) ExtractSPKIHashExported(certBytes []byte) ([]byte, error) {
	return k.extractSPKIHash(certBytes)
}

// UpdateRandomSourceStatusExported exports updateRandomSourceStatus for testing
func (k Keeper) UpdateRandomSourceStatusExported(
	ctx context.Context,
	sourceID string,
	status cryptoproto.RandomSourceStatus,
) error {
	return k.updateRandomSourceStatus(ctx, sourceID, status)
}
