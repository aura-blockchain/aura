// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package app

import tmlog "cosmossdk.io/log"

// CosmosApp aliases App to preserve backwards compatibility with existing tests.
type CosmosApp = App

// NewCosmosApp creates a fully wired Cosmos SDK application with the provided logger.
func NewCosmosApp(logger tmlog.Logger) *CosmosApp {
	return NewAppWithOptions(logger, nil, "")
}
