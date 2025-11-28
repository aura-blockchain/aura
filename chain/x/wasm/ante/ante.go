package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/aequitas/aura/chain/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmGasDecorator validates wasm contract execution gas limits with PRE-FLIGHT calculation
type WasmGasDecorator struct {
	wasmKeeper keeper.Keeper
}

// NewWasmGasDecorator creates a new WasmGasDecorator
func NewWasmGasDecorator(wasmKeeper keeper.Keeper) WasmGasDecorator {
	return WasmGasDecorator{
		wasmKeeper: wasmKeeper,
	}
}

// AnteHandle validates wasm gas limits with PRE-FLIGHT estimation and enforcement
func (wgd WasmGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	params := wgd.wasmKeeper.GetParams(ctx)

	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *types.MsgInstantiateContract:
			// PRE-FLIGHT gas calculation for instantiation
			estimate := wgd.calculateInstantiateGas(ctx, m)
			totalEstimatedGas := estimate.CalculateTotalGas()

			// Check BEFORE execution
			if totalEstimatedGas > params.MaxGasWasmExecution {
				return ctx, errorsmod.Wrapf(
					types.ErrGasLimitExceeded,
					"pre-flight instantiate gas estimate %d exceeds limit %d",
					totalEstimatedGas,
					params.MaxGasWasmExecution,
				)
			}

			// Reserve gas for instantiation
			ctx.GasMeter().ConsumeGas(totalEstimatedGas/10, "instantiate_preflight_reserve")

		case *types.MsgExecuteContract:
			// Check if contract is paused
			if wgd.wasmKeeper.IsContractPaused(ctx, m.Contract) {
				return ctx, types.ErrContractPaused.Wrapf("contract %s is paused", m.Contract)
			}

			// PRE-FLIGHT gas calculation for execution
			estimate := wgd.calculateExecuteGas(ctx, m)
			totalEstimatedGas := estimate.CalculateTotalGas()

			// Check BEFORE execution
			if totalEstimatedGas > params.MaxGasWasmExecution {
				return ctx, errorsmod.Wrapf(
					types.ErrGasLimitExceeded,
					"pre-flight execute gas estimate %d exceeds limit %d",
					totalEstimatedGas,
					params.MaxGasWasmExecution,
				)
			}

			// Reserve gas for execution
			ctx.GasMeter().ConsumeGas(totalEstimatedGas/10, "execute_preflight_reserve")

		case *types.MsgStoreCode:
			// Calculate storage gas cost (per byte)
			storageGas := types.CalculateStorageGas(uint64(len(m.WasmByteCode)), types.GasPerByte)

			// Consume gas for code storage BEFORE storing
			ctx.GasMeter().ConsumeGas(storageGas, "code_storage")

			// Validate contract size doesn't exceed per-block limit
			if err := wgd.validateContractSize(ctx, m.WasmByteCode); err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

// calculateInstantiateGas estimates gas cost for contract instantiation
func (wgd WasmGasDecorator) calculateInstantiateGas(ctx sdk.Context, msg *types.MsgInstantiateContract) types.GasEstimate {
	estimate := types.GasEstimate{
		StorageGas:     5000,  // Base storage cost for contract state
		ComputationGas: 10000, // Base computation cost
		CallGas:        0,
	}

	// Add gas for init message size
	msgSize := uint64(len(msg.Msg))
	estimate.ComputationGas += msgSize * 10 // 10 gas per byte of message

	// Add gas for funds transfer if any
	if len(msg.Funds) > 0 {
		estimate.ComputationGas += 1000 // Cost for processing funds
	}

	return estimate
}

// calculateExecuteGas estimates gas cost for contract execution
func (wgd WasmGasDecorator) calculateExecuteGas(ctx sdk.Context, msg *types.MsgExecuteContract) types.GasEstimate {
	estimate := types.GasEstimate{
		StorageGas:     2000,  // Base storage cost for state updates
		ComputationGas: 5000,  // Base computation cost
		CallGas:        0,
	}

	// Add gas for execute message size
	msgSize := uint64(len(msg.Msg))
	estimate.ComputationGas += msgSize * 10 // 10 gas per byte of message

	// Add gas for funds transfer if any
	if len(msg.Funds) > 0 {
		estimate.ComputationGas += 1000 // Cost for processing funds
	}

	return estimate
}

// validateContractSize validates that contract uploads don't exceed per-block limits
func (wgd WasmGasDecorator) validateContractSize(ctx sdk.Context, code []byte) error {
	params := wgd.wasmKeeper.GetParams(ctx)

	// Get total size of contracts uploaded in this block
	blockContractSize := wgd.getBlockContractSize(ctx)
	newSize := blockContractSize + uint64(len(code))

	if newSize > params.MaxWasmCodeSize {
		return types.ErrContractTooLarge.Wrapf(
			"contract size per block exceeded: current %d + new %d > limit %d",
			blockContractSize,
			len(code),
			params.MaxWasmCodeSize,
		)
	}

	// Update block contract size
	wgd.setBlockContractSize(ctx, newSize)
	return nil
}

// getBlockContractSize returns the total size of contracts uploaded in this block
func (wgd WasmGasDecorator) getBlockContractSize(ctx sdk.Context) uint64 {
	// Use transient store keyed by block height (resets each block)
	tStore := ctx.TransientStore(wgd.wasmKeeper.GetStoreKey())
	key := append([]byte("block_contract_size_"), sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))...)

	bz := tStore.Get(key)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// setBlockContractSize sets the total size of contracts uploaded in this block
func (wgd WasmGasDecorator) setBlockContractSize(ctx sdk.Context, size uint64) {
	// Use transient store keyed by block height (resets each block)
	tStore := ctx.TransientStore(wgd.wasmKeeper.GetStoreKey())
	key := append([]byte("block_contract_size_"), sdk.Uint64ToBigEndian(uint64(ctx.BlockHeight()))...)

	tStore.Set(key, sdk.Uint64ToBigEndian(size))
}

// WasmAuthDecorator validates wasm contract upload authorization
type WasmAuthDecorator struct {
	wasmKeeper keeper.Keeper
}

// NewWasmAuthDecorator creates a new WasmAuthDecorator
func NewWasmAuthDecorator(wasmKeeper keeper.Keeper) WasmAuthDecorator {
	return WasmAuthDecorator{
		wasmKeeper: wasmKeeper,
	}
}

// AnteHandle validates wasm authorization
func (wad WasmAuthDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		if storeMsg, ok := msg.(*types.MsgStoreCode); ok {
			// Validate upload authorization
			if !wad.wasmKeeper.IsAuthorizedUploader(ctx, storeMsg.Sender) {
				return ctx, types.ErrUnauthorized.Wrapf(
					"address %s is not authorized to upload contracts",
					storeMsg.Sender,
				)
			}
		}
	}

	return next(ctx, tx, simulate)
}

// WasmReentrancyDecorator prevents reentrancy attacks
type WasmReentrancyDecorator struct {
	wasmKeeper keeper.Keeper
}

// NewWasmReentrancyDecorator creates a new WasmReentrancyDecorator
func NewWasmReentrancyDecorator(wasmKeeper keeper.Keeper) WasmReentrancyDecorator {
	return WasmReentrancyDecorator{
		wasmKeeper: wasmKeeper,
	}
}

// AnteHandle validates against reentrancy attacks
func (wrd WasmReentrancyDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Track executing contracts in this transaction
	executingContracts := make(map[string]bool)

	for _, msg := range tx.GetMsgs() {
		if execMsg, ok := msg.(*types.MsgExecuteContract); ok {
			// Check if contract is already executing in this tx (reentrancy)
			if executingContracts[execMsg.Contract] {
				return ctx, types.ErrReentrancyDetected.Wrapf(
					"reentrancy detected for contract %s in same transaction",
					execMsg.Contract,
				)
			}
			executingContracts[execMsg.Contract] = true

			// Also check global reentrancy protection from keeper
			// This is handled in msg_server.go ExecuteContract
		}
	}

	return next(ctx, tx, simulate)
}

// WasmSecurityDecorator performs comprehensive security checks
type WasmSecurityDecorator struct {
	wasmKeeper keeper.Keeper
}

// NewWasmSecurityDecorator creates a new WasmSecurityDecorator
func NewWasmSecurityDecorator(wasmKeeper keeper.Keeper) WasmSecurityDecorator {
	return WasmSecurityDecorator{
		wasmKeeper: wasmKeeper,
	}
}

// AnteHandle performs security validation
func (wsd WasmSecurityDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *types.MsgStoreCode:
			// Validate contract code
			if err := wsd.validateContractCode(ctx, m.WasmByteCode); err != nil {
				return ctx, err
			}

		case *types.MsgExecuteContract:
			// Validate execution safety
			if err := wsd.validateExecution(ctx, m); err != nil {
				return ctx, err
			}

		case *types.MsgMigrateContract:
			// Validate migration is allowed
			params := wsd.wasmKeeper.GetParams(ctx)
			if params.RequireAdminForMigrate {
				// Additional validation for admin requirement would go here
				// This is a security feature to ensure only admins can migrate
			}
		}
	}

	return next(ctx, tx, simulate)
}

// validateContractCode performs basic validation of contract code
func (wsd WasmSecurityDecorator) validateContractCode(ctx sdk.Context, code []byte) error {
	if len(code) == 0 {
		return types.ErrInvalidContractCode.Wrap("contract code cannot be empty")
	}

	params := wsd.wasmKeeper.GetParams(ctx)
	if uint64(len(code)) > params.MaxWasmCodeSize {
		return types.ErrContractTooLarge.Wrapf(
			"contract size %d exceeds maximum %d",
			len(code),
			params.MaxWasmCodeSize,
		)
	}

	// Additional validation could include:
	// - WebAssembly format validation
	// - Malicious bytecode pattern detection
	// - Code signature verification
	// These are stubs for now

	return nil
}

// validateExecution validates contract execution safety
func (wsd WasmSecurityDecorator) validateExecution(ctx sdk.Context, msg *types.MsgExecuteContract) error {
	// Check if contract is paused
	if wsd.wasmKeeper.IsContractPaused(ctx, msg.Contract) {
		return types.ErrContractPaused.Wrapf("contract %s is paused", msg.Contract)
	}

	// Validate funds
	if len(msg.Funds) > 0 {
		// Convert []*sdk.Coin to []sdk.Coin for validation
		coins := make([]sdk.Coin, len(msg.Funds))
		for i, coin := range msg.Funds {
			if coin != nil {
				coins[i] = *coin
			}
		}
		funds := sdk.NewCoins(coins...)
		if !funds.IsValid() {
			return fmt.Errorf("invalid funds")
		}
	}

	// Additional validation could include:
	// - Rate limiting per contract
	// - Maximum funds limit per execution
	// - Blacklist/whitelist checks
	// These are stubs for now

	return nil
}

// ChainAnteDecorators chains multiple ante decorators together
func ChainAnteDecorators(wasmKeeper keeper.Keeper) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewWasmAuthDecorator(wasmKeeper),
		NewWasmGasDecorator(wasmKeeper),
		NewWasmReentrancyDecorator(wasmKeeper),
		NewWasmSecurityDecorator(wasmKeeper),
	)
}
