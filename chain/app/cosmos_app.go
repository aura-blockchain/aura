package app

import tmlog "cosmossdk.io/log"

// CosmosApp aliases App to preserve backwards compatibility with existing tests.
type CosmosApp = App

// NewCosmosApp mirrors the old constructor while reusing the fully wired application.
func NewCosmosApp(logger tmlog.Logger) *CosmosApp {
	return NewAppWithLogger(logger)
}
