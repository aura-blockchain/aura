// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the module name for compliance.
	ModuleName = "compliance"

	// StoreKey defines the primary KV store key.
	StoreKey = ModuleName

	// RouterKey is used for routing compliance messages.
	RouterKey = ModuleName

	// QuerierRoute defines the query routing key.
	QuerierRoute = ModuleName

	// MemStoreKey exposes the in-memory store key (unused but reserved).
	MemStoreKey = "mem_" + ModuleName
)
