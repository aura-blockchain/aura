// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewChaosEngine(t *testing.T) {
	config := DefaultChaosConfig()
	engine := NewChaosEngine(config)

	require.NotNil(t, engine)
	require.True(t, engine.active)
	require.NotNil(t, engine.config)
}

func TestShouldFail(t *testing.T) {
	config := DefaultChaosConfig()
	config.FailureRate = 1.0 // Always fail
	engine := NewChaosEngine(config)

	require.True(t, engine.ShouldFail())

	config.FailureRate = 0.0 // Never fail
	engine = NewChaosEngine(config)
	require.False(t, engine.ShouldFail())
}

func TestInjectLatency(t *testing.T) {
	config := DefaultChaosConfig()
	config.MaxLatencyMs = 100
	engine := NewChaosEngine(config)

	latency := engine.InjectLatency()
	require.GreaterOrEqual(t, latency, time.Duration(0))
	require.LessOrEqual(t, latency, 100*time.Millisecond)
}

func TestInjectDataCorruption(t *testing.T) {
	config := DefaultChaosConfig()
	config.DataCorruptionProb = 1.0 // Always corrupt
	engine := NewChaosEngine(config)

	data := []byte{0x01, 0x02, 0x03, 0x04}
	corrupted := engine.InjectDataCorruption(data)

	// At least one byte should be different
	different := false
	for i := range data {
		if data[i] != corrupted[i] {
			different = true
			break
		}
	}
	require.True(t, different)
}

func TestChaosEngineStart(t *testing.T) {
	config := DefaultChaosConfig()
	engine := NewChaosEngine(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	engine.Start(ctx)
	time.Sleep(1 * time.Second)
	engine.Stop()

	require.False(t, engine.active)
}

func TestGetInjections(t *testing.T) {
	config := DefaultChaosConfig()
	engine := NewChaosEngine(config)

	// Record some injections
	engine.InjectLatency()
	engine.InjectSlowdown()

	injections := engine.GetInjections()
	require.NotNil(t, injections)
}

func TestPrintReport(t *testing.T) {
	config := DefaultChaosConfig()
	engine := NewChaosEngine(config)

	// Record some injections
	engine.InjectLatency()
	engine.InjectSlowdown()

	// Should not panic
	engine.PrintReport()
}

func TestChaosDisabled(t *testing.T) {
	config := DefaultChaosConfig()
	config.Enabled = false
	engine := NewChaosEngine(config)

	require.False(t, engine.ShouldFail())
	require.Equal(t, time.Duration(0), engine.InjectLatency())

	data := []byte{0x01, 0x02, 0x03}
	corrupted := engine.InjectDataCorruption(data)
	require.Equal(t, data, corrupted)
}
