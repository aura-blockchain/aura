// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package types

const (
	// ModuleName defines the module name
	ModuleName = "aurabindings"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_" + ModuleName
)

// KVStore key prefixes
var (
	// QueryStatsPrefix is the prefix for query statistics
	QueryStatsPrefix = []byte{0x01}

	// MessageStatsPrefix is the prefix for message statistics
	MessageStatsPrefix = []byte{0x02}

	// RateLimitPrefix is the prefix for rate limit data
	RateLimitPrefix = []byte{0x03}
)

// Constants for rate limiting and validation
const (
	// MaxQueriesPerBlock is the maximum number of queries allowed per address per block
	MaxQueriesPerBlock = 1000

	// MaxMessagesPerBlock is the maximum number of messages allowed per address per block
	MaxMessagesPerBlock = 100
)
