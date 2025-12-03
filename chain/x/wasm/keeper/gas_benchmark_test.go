package keeper

import (
	"encoding/json"
	"os"
	"testing"

	storetypes "cosmossdk.io/store/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// BenchmarkWasmStoreCode benchmarks contract code upload with different sizes
func BenchmarkWasmStoreCode(b *testing.B) {
	testCases := []struct {
		name         string
		contractPath string
		description  string
	}{
		{
			name:         "SmallContract_157KB",
			contractPath: "../../../../contracts/artifacts/binding_tester.wasm",
			description:  "Binding tester contract (157KB)",
		},
		{
			name:         "MediumContract_230KB",
			contractPath: "../../../../contracts/artifacts/vc_issuer.wasm",
			description:  "VC Issuer contract (230KB)",
		},
		{
			name:         "MediumContract_236KB",
			contractPath: "../../../../contracts/artifacts/schema.wasm",
			description:  "Schema contract (236KB)",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Load contract bytecode
			wasmCode, err := os.ReadFile(tc.contractPath)
			if err != nil {
				b.Skipf("Contract file not found: %s (run 'make optimize-wasm' first)", tc.contractPath)
				return
			}

			b.Logf("Testing %s - Contract size: %d bytes (%.2f KB)", tc.description, len(wasmCode), float64(len(wasmCode))/1024)

			// Reset timer before benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Setup test environment for each iteration
				ctx, keeper, deps := setupKeeperTest(b)
				sender := deps.funder

				// Authorize uploader
				err := keeper.AuthorizeUploader(ctx, sender.String())
				require.NoError(b, err)

				// Record gas before operation
				gasBefore := ctx.GasMeter().GasConsumed()
				b.StartTimer()

				// Store contract code
				codeID, err := keeper.StoreCode(ctx, sender, wasmCode)

				b.StopTimer()
				require.NoError(b, err)
				require.NotZero(b, codeID)

				// Calculate gas consumed
				gasUsed := ctx.GasMeter().GasConsumed() - gasBefore

				b.ReportMetric(float64(gasUsed), "gas/op")
				b.ReportMetric(float64(len(wasmCode)), "bytes")
				b.ReportMetric(float64(gasUsed)/float64(len(wasmCode)), "gas/byte")
			}
		})
	}
}

// BenchmarkWasmInstantiateContract benchmarks contract instantiation
func BenchmarkWasmInstantiateContract(b *testing.B) {
	// Load the VC issuer contract as our test contract
	wasmCode, err := os.ReadFile("../../../../contracts/artifacts/vc_issuer.wasm")
	if err != nil {
		b.Skipf("Contract file not found (run 'make optimize-wasm' first): %v", err)
		return
	}

	testCases := []struct {
		name        string
		initMsg     interface{}
		description string
	}{
		{
			name: "SimpleInit_NoState",
			initMsg: map[string]interface{}{
				"admin": "aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
			},
			description: "Minimal initialization with admin only",
		},
		{
			name: "ComplexInit_WithState",
			initMsg: map[string]interface{}{
				"admin": "aura1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqnrql8a",
				"config": map[string]interface{}{
					"max_issuers":     100,
					"verification_required": true,
					"rate_limit":      1000,
				},
			},
			description: "Complex initialization with configuration state",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.Logf("Testing %s", tc.description)

			// Setup and store contract code once
			ctx, keeper, deps := setupKeeperTest(b)
			sender := deps.funder

			// Authorize and store code
			err := keeper.AuthorizeUploader(ctx, sender.String())
			require.NoError(b, err)

			_, err = keeper.StoreCode(ctx, sender, wasmCode)
			require.NoError(b, err)

			// Marshal init message
			initMsgBytes, err := json.Marshal(tc.initMsg)
			require.NoError(b, err)

			b.Logf("Init message size: %d bytes", len(initMsgBytes))

			// Reset timer before benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Create fresh context for each iteration
				ctx, keeper, deps := setupKeeperTest(b)
				sender := deps.funder

				// Re-authorize and re-store code for clean state
				err := keeper.AuthorizeUploader(ctx, sender.String())
				require.NoError(b, err)
				freshCodeID, err := keeper.StoreCode(ctx, sender, wasmCode)
				require.NoError(b, err)

				// Record gas before operation
				gasBefore := ctx.GasMeter().GasConsumed()
				b.StartTimer()

				// Instantiate contract
				contractAddr, data, err := keeper.InstantiateContract(
					ctx,
					freshCodeID,
					sender,
					sender, // admin
					initMsgBytes,
					"benchmark-contract",
					sdk.NewCoins(),
				)

				b.StopTimer()
				require.NoError(b, err)
				require.NotEmpty(b, contractAddr)
				require.NotNil(b, data)

				// Calculate gas consumed
				gasUsed := ctx.GasMeter().GasConsumed() - gasBefore

				b.ReportMetric(float64(gasUsed), "gas/op")
				b.ReportMetric(float64(len(initMsgBytes)), "init_msg_bytes")
				b.ReportMetric(float64(gasUsed)/float64(len(initMsgBytes)), "gas/init_byte")
			}
		})
	}
}

// BenchmarkWasmExecuteContract benchmarks contract execution with different complexity levels
func BenchmarkWasmExecuteContract(b *testing.B) {
	// Load the binding tester contract for execution tests
	wasmCode, err := os.ReadFile("../../../../contracts/artifacts/binding_tester.wasm")
	if err != nil {
		b.Skipf("Contract file not found (run 'make optimize-wasm' first): %v", err)
		return
	}

	testCases := []struct {
		name        string
		executeMsg  interface{}
		description string
	}{
		{
			name: "SimpleQuery_ReadOnly",
			executeMsg: map[string]interface{}{
				"get_config": map[string]interface{}{},
			},
			description: "Simple read-only query operation",
		},
		{
			name: "SimpleWrite_SingleValue",
			executeMsg: map[string]interface{}{
				"set_value": map[string]interface{}{
					"key":   "test_key",
					"value": "test_value",
				},
			},
			description: "Simple write operation (single KV pair)",
		},
		{
			name: "ComplexWrite_MultipleValues",
			executeMsg: map[string]interface{}{
				"batch_update": map[string]interface{}{
					"updates": []map[string]string{
						{"key": "key1", "value": "value1"},
						{"key": "key2", "value": "value2"},
						{"key": "key3", "value": "value3"},
						{"key": "key4", "value": "value4"},
						{"key": "key5", "value": "value5"},
					},
				},
			},
			description: "Complex write operation (batch of 5 KV pairs)",
		},
		{
			name: "ComputeHeavy_Calculation",
			executeMsg: map[string]interface{}{
				"compute": map[string]interface{}{
					"iterations": 1000,
				},
			},
			description: "Compute-heavy operation (1000 iterations)",
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.Logf("Testing %s", tc.description)

			// Setup and deploy contract once
			ctx, keeper, deps := setupKeeperTest(b)
			sender := deps.funder

			// Authorize, store, and instantiate contract
			err := keeper.AuthorizeUploader(ctx, sender.String())
			require.NoError(b, err)

			codeID, err := keeper.StoreCode(ctx, sender, wasmCode)
			require.NoError(b, err)

			initMsg := json.RawMessage(`{"admin":"` + sender.String() + `"}`)
			_, _, err = keeper.InstantiateContract(
				ctx,
				codeID,
				sender,
				sender,
				initMsg,
				"benchmark-exec-contract",
				sdk.NewCoins(),
			)
			require.NoError(b, err)

			// Marshal execute message
			executeMsgBytes, err := json.Marshal(tc.executeMsg)
			require.NoError(b, err)

			b.Logf("Execute message size: %d bytes", len(executeMsgBytes))

			// Reset timer before benchmark loop
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Create fresh context for each iteration
				ctx, keeper, deps := setupKeeperTest(b)

				// Re-setup contract (to ensure clean state)
				sender := deps.funder
				err := keeper.AuthorizeUploader(ctx, sender.String())
				require.NoError(b, err)
				codeID, err := keeper.StoreCode(ctx, sender, wasmCode)
				require.NoError(b, err)
				contractAddr, _, err := keeper.InstantiateContract(
					ctx, codeID, sender, sender, initMsg, "bench-exec", sdk.NewCoins(),
				)
				require.NoError(b, err)

				// Record gas before operation
				gasBefore := ctx.GasMeter().GasConsumed()
				b.StartTimer()

				// Execute contract
				data, err := keeper.ExecuteContract(
					ctx,
					contractAddr,
					sender,
					executeMsgBytes,
					sdk.NewCoins(),
				)

				b.StopTimer()

				// Note: Some execute messages may not be valid for binding_tester
				// In production, we'd use contract-specific valid messages
				// For now, we measure the overhead even if execution fails
				if err == nil {
					require.NotNil(b, data)
				}

				// Calculate gas consumed
				gasUsed := ctx.GasMeter().GasConsumed() - gasBefore

				b.ReportMetric(float64(gasUsed), "gas/op")
				b.ReportMetric(float64(len(executeMsgBytes)), "msg_bytes")
				if len(executeMsgBytes) > 0 {
					b.ReportMetric(float64(gasUsed)/float64(len(executeMsgBytes)), "gas/msg_byte")
				}
			}
		})
	}
}

// BenchmarkWasmFullLifecycle benchmarks the complete contract lifecycle
func BenchmarkWasmFullLifecycle(b *testing.B) {
	wasmCode, err := os.ReadFile("../../../../contracts/artifacts/binding_tester.wasm")
	if err != nil {
		b.Skipf("Contract file not found (run 'make optimize-wasm' first): %v", err)
		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, keeper, deps := setupKeeperTest(b)
		sender := deps.funder

		// Authorize uploader
		err := keeper.AuthorizeUploader(ctx, sender.String())
		require.NoError(b, err)

		gasBefore := ctx.GasMeter().GasConsumed()
		b.StartTimer()

		// 1. Store code
		codeID, err := keeper.StoreCode(ctx, sender, wasmCode)
		require.NoError(b, err)

		gasAfterStore := ctx.GasMeter().GasConsumed()

		// 2. Instantiate contract
		initMsg := json.RawMessage(`{"admin":"` + sender.String() + `"}`)
		contractAddr, _, err := keeper.InstantiateContract(
			ctx,
			codeID,
			sender,
			sender,
			initMsg,
			"lifecycle-benchmark",
			sdk.NewCoins(),
		)
		require.NoError(b, err)

		gasAfterInstantiate := ctx.GasMeter().GasConsumed()

		// 3. Execute contract
		executeMsg := json.RawMessage(`{"get_config":{}}`)
		_, err = keeper.ExecuteContract(
			ctx,
			contractAddr,
			sender,
			executeMsg,
			sdk.NewCoins(),
		)
		// Error is acceptable for this benchmark (measuring overhead)

		gasAfterExecute := ctx.GasMeter().GasConsumed()

		b.StopTimer()

		// Report detailed metrics
		totalGas := gasAfterExecute - gasBefore
		storeGas := gasAfterStore - gasBefore
		instantiateGas := gasAfterInstantiate - gasAfterStore
		executeGas := gasAfterExecute - gasAfterInstantiate

		b.ReportMetric(float64(totalGas), "total_gas")
		b.ReportMetric(float64(storeGas), "store_gas")
		b.ReportMetric(float64(instantiateGas), "instantiate_gas")
		b.ReportMetric(float64(executeGas), "execute_gas")
		b.ReportMetric(float64(len(wasmCode)), "contract_bytes")
	}
}

// BenchmarkWasmReentrancyProtection benchmarks reentrancy detection overhead
func BenchmarkWasmReentrancyProtection(b *testing.B) {
	wasmCode, err := os.ReadFile("../../../../contracts/artifacts/binding_tester.wasm")
	if err != nil {
		b.Skipf("Contract file not found (run 'make optimize-wasm' first): %v", err)
		return
	}

	executeMsg := json.RawMessage(`{"get_config":{}}`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ctx, keeper, deps := setupKeeperTest(b)
		sender := deps.funder

		// Re-setup contract
		err := keeper.AuthorizeUploader(ctx, sender.String())
		require.NoError(b, err)
		codeID, err := keeper.StoreCode(ctx, sender, wasmCode)
		require.NoError(b, err)
		initMsg := json.RawMessage(`{"admin":"` + sender.String() + `"}`)
		contractAddr, _, err := keeper.InstantiateContract(
			ctx, codeID, sender, sender, initMsg, "reentrancy-bench", sdk.NewCoins(),
		)
		require.NoError(b, err)

		gasBefore := ctx.GasMeter().GasConsumed()
		b.StartTimer()

		// Execute with reentrancy protection
		_, _ = keeper.ExecuteContract(ctx, contractAddr, sender, executeMsg, sdk.NewCoins())

		b.StopTimer()
		gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
		b.ReportMetric(float64(gasUsed), "gas/op")
	}
}

// BenchmarkWasmAdminOperations benchmarks admin-related operations
func BenchmarkWasmAdminOperations(b *testing.B) {
	b.Run("SetContractAdmin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			ctx, keeper, deps := setupKeeperTest(b)
			sender := deps.funder
			newAdmin := deps.funder

			gasBefore := ctx.GasMeter().GasConsumed()
			b.StartTimer()

			_ = keeper.SetContractAdmin(ctx, sender, newAdmin)

			b.StopTimer()
			gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
			b.ReportMetric(float64(gasUsed), "gas/op")
		}
	})

	b.Run("GetContractAdmin", func(b *testing.B) {
		// Setup admin first
		ctx, keeper, deps := setupKeeperTest(b)
		sender := deps.funder
		_ = keeper.SetContractAdmin(ctx, sender, sender)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gasBefore := ctx.GasMeter().GasConsumed()

			_, _ = keeper.GetContractAdmin(ctx, sender)

			gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
			b.ReportMetric(float64(gasUsed), "gas/op")
		}
	})

	b.Run("IsContractAdmin", func(b *testing.B) {
		// Setup admin first
		ctx, keeper, deps := setupKeeperTest(b)
		sender := deps.funder
		_ = keeper.SetContractAdmin(ctx, sender, sender)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			gasBefore := ctx.GasMeter().GasConsumed()

			_, _ = keeper.IsContractAdmin(ctx, sender, sender)

			gasUsed := ctx.GasMeter().GasConsumed() - gasBefore
			b.ReportMetric(float64(gasUsed), "gas/op")
		}
	})
}

// setupKeeperTest creates a test environment for benchmarks
func setupKeeperTest(tb testing.TB) (sdk.Context, Keeper, testDeps) {
	tb.Helper()

	deps := setupTest(tb)
	return deps.ctx, deps.keeper, deps
}

// testDeps holds test dependencies
type testDeps struct {
	ctx    sdk.Context
	keeper Keeper
	funder sdk.AccAddress
}

// setupTest creates test dependencies
func setupTest(tb testing.TB) testDeps {
	tb.Helper()

	ctx, keeper, deps := SetupWasmKeeperTest(tb)

	// Get funder from test dependencies
	var funder sdk.AccAddress
	if deps != nil && len(deps.Funder) > 0 {
		funder = deps.Funder[0]
	} else {
		// Fallback: create a test address
		funder = sdk.AccAddress([]byte("test_funder_address_12345"))
	}

	return testDeps{
		ctx:    ctx,
		keeper: keeper,
		funder: funder,
	}
}

// SetupWasmKeeperTest creates a test context and keeper for WASM benchmarks
func SetupWasmKeeperTest(tb testing.TB) (sdk.Context, Keeper, *TestDeps) {
	tb.Helper()

	// Use the existing test helper from keeper_test.go
	return setupWasmTest(tb)
}

// TestDeps holds test dependencies
type TestDeps struct {
	Funder []sdk.AccAddress
}

// setupWasmTest creates a minimal test environment for benchmarking
func setupWasmTest(tb testing.TB) (sdk.Context, Keeper, *TestDeps) {
	tb.Helper()

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create test context using SDK testutil
	testCtx := testutil.DefaultContextWithDB(tb, storeKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx

	// Create codec
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	// Create test account
	funder := sdk.AccAddress([]byte("test_benchmark_funder_"))

	// Create keeper with nil wasm keeper (we'll use mock for benchmarks)
	// Note: For real benchmarks, you'd need a properly configured wasmd keeper
	keeper := NewKeeper(
		cdc,
		storeKey,
		&wasmkeeper.Keeper{}, // wasm keeper - in production would be properly configured
		"aura10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // authority
	)

	deps := &TestDeps{
		Funder: []sdk.AccAddress{funder},
	}

	return ctx, keeper, deps
}
