package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ============================================================================
// PARAMETERS
// ============================================================================

// GetParams returns the module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return *types.DefaultParams()
	}

	var params types.Params
	if err := json.Unmarshal(bz, &params); err != nil {
		return *types.DefaultParams()
	}
	return params
}

// SetParams sets the module parameters
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := types.ValidateParams(&params); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// UpdateParams updates the module parameters (governance only)
func (k Keeper) UpdateParams(ctx context.Context, authority string, params types.Params) error {
	if authority != k.authority {
		return types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", k.authority, authority)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return k.SetParams(sdkCtx, params)
}

// ============================================================================
// AUTHORIZATION
// ============================================================================

// IsAuthorizedUploader checks if an address is authorized to upload contracts
func (k Keeper) IsAuthorizedUploader(ctx sdk.Context, address string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAuthKey(address)
	return store.Has(key)
}

// AuthorizeUploader authorizes an address to upload contracts
func (k Keeper) AuthorizeUploader(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAuthKey(address)
	store.Set(key, []byte{0x01})
	return nil
}

// RevokeUploader revokes contract upload authorization
func (k Keeper) RevokeUploader(ctx sdk.Context, address string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractAuthKey(address)
	store.Delete(key)
	return nil
}

// ============================================================================
// CONTRACT PAUSE/UNPAUSE
// ============================================================================

// IsContractPaused checks if a contract is paused
func (k Keeper) IsContractPaused(ctx sdk.Context, contractAddr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractPauseKey(contractAddr)
	return store.Has(key)
}

// PauseContract pauses a contract
func (k Keeper) PauseContract(ctx sdk.Context, contractAddr string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractPauseKey(contractAddr)
	store.Set(key, []byte{0x01})

	// Update stats
	k.incrementSecurityStat(ctx, "paused")

	return nil
}

// UnpauseContract unpauses a contract
func (k Keeper) UnpauseContract(ctx sdk.Context, contractAddr string) error {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractPauseKey(contractAddr)
	if !store.Has(key) {
		return types.ErrContractPaused.Wrapf("contract %s is not paused", contractAddr)
	}
	store.Delete(key)

	// Update stats
	k.decrementSecurityStat(ctx, "paused")

	return nil
}

// ============================================================================
// CONTRACT OPERATIONS
// ============================================================================

// ValidateContractUpload validates contract upload authorization and code
func (k Keeper) ValidateContractUpload(ctx sdk.Context, sender string, code []byte) error {
	// Check authorization
	if !k.IsAuthorizedUploader(ctx, sender) {
		params := k.GetParams(ctx)
		if params.CodeUploadAccess.Permission != types.AccessTypeEverybody {
			return types.ErrUnauthorized.Wrapf("address %s is not authorized to upload contracts", sender)
		}
	}

	// Validate code
	if len(code) == 0 {
		return types.ErrInvalidContractCode.Wrap("contract code cannot be empty")
	}

	params := k.GetParams(ctx)
	if uint64(len(code)) > params.MaxWasmCodeSize {
		return types.ErrContractTooLarge.Wrapf("contract size %d exceeds maximum %d", len(code), params.MaxWasmCodeSize)
	}

	return nil
}

// ValidateContractExecution validates that a contract can be executed
func (k Keeper) ValidateContractExecution(ctx sdk.Context, contractAddr string) error {
	// Check if contract is paused
	if k.IsContractPaused(ctx, contractAddr) {
		return types.ErrContractPaused.Wrapf("contract %s is paused", contractAddr)
	}

	// Additional validation could include:
	// - Check if contract exists
	// - Check if contract is not blacklisted
	// - Check if contract has not exceeded rate limits
	// For now, just check pause status

	return nil
}

// StoreCode stores contract code
func (k Keeper) StoreCode(ctx sdk.Context, sender sdk.AccAddress, code []byte) (uint64, error) {
	if k.wasmKeeper == nil {
		return 0, fmt.Errorf("wasm keeper not configured")
	}

	ops := wasmkeeper.NewDefaultPermissionKeeper(k.wasmKeeper)
	codeID, _, err := ops.Create(ctx, sender, code, nil)
	if err != nil {
		return 0, err
	}

	// Update stats
	k.incrementSecurityStat(ctx, "contracts_uploaded")

	return codeID, nil
}

// InstantiateContract instantiates a contract
func (k Keeper) InstantiateContract(
	ctx sdk.Context,
	codeID uint64,
	creator sdk.AccAddress,
	admin sdk.AccAddress,
	initMsg []byte,
	label string,
	funds sdk.Coins,
) (sdk.AccAddress, []byte, error) {
	// Call before hook
	if err := k.BeforeInstantiateHook(ctx, codeID, creator, admin, label); err != nil {
		return nil, nil, err
	}

	if k.wasmKeeper == nil {
		return nil, nil, fmt.Errorf("wasm keeper not configured")
	}

	ops := wasmkeeper.NewDefaultPermissionKeeper(k.wasmKeeper)
	contractAddr, data, err := ops.Instantiate(ctx, codeID, creator, admin, initMsg, label, funds)
	if err != nil {
		return nil, nil, err
	}

	// Call after hook
	if err := k.AfterInstantiateHook(ctx, contractAddr, codeID, creator, admin, label); err != nil {
		// Log error but don't fail instantiation
		k.Logger(ctx).Error("after instantiate hook failed", "error", err)
	}

	return contractAddr, data, nil
}

// ExecuteContract executes a contract
func (k Keeper) ExecuteContract(ctx sdk.Context, contractAddr, sender sdk.AccAddress, msg []byte, funds sdk.Coins) ([]byte, error) {
	// Check if contract is paused
	if k.IsContractPaused(ctx, contractAddr.String()) {
		return nil, types.ErrContractPaused.Wrapf("contract %s is paused", contractAddr.String())
	}

	// Call before hook
	if err := k.BeforeExecuteHook(ctx, contractAddr, sender); err != nil {
		return nil, err
	}

	if k.wasmKeeper == nil {
		return nil, fmt.Errorf("wasm keeper not configured")
	}
	ops := wasmkeeper.NewDefaultPermissionKeeper(k.wasmKeeper)

	gasBefore := ctx.GasMeter().GasConsumed()
	data, err := ops.Execute(ctx, contractAddr, sender, msg, funds)
	if err != nil {
		return nil, err
	}

	// Call after hook
	success := true
	k.AfterExecuteHook(ctx, contractAddr, gasBefore, success, nil)

	return data, nil
}

// Migrate migrates a contract to new code
func (k Keeper) Migrate(ctx sdk.Context, contractAddr, caller sdk.AccAddress, newCodeID uint64, msg []byte) ([]byte, error) {
	// Check migration requirements
	params := k.GetParams(ctx)
	if params.RequireAdminForMigrate {
		// In a full implementation, would check if caller is admin of contract
		// For now, just note the requirement is enabled
	}

	if k.wasmKeeper == nil {
		return nil, fmt.Errorf("wasm keeper not configured")
	}

	ops := wasmkeeper.NewDefaultPermissionKeeper(k.wasmKeeper)
	data, err := ops.Migrate(ctx, contractAddr, caller, newCodeID, msg)
	if err != nil {
		return nil, err
	}

	// Update stats
	k.incrementSecurityStat(ctx, "contracts_migrated")

	return data, nil
}

// QuerySmart queries a smart contract
func (k Keeper) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, queryMsg []byte) ([]byte, error) {
	// Check if contract is paused
	if k.IsContractPaused(ctx, contractAddr.String()) {
		return nil, types.ErrContractPaused.Wrapf("contract %s is paused", contractAddr.String())
	}

	if k.wasmKeeper == nil {
		return nil, fmt.Errorf("wasm keeper not configured")
	}

	return k.wasmKeeper.QuerySmart(ctx, contractAddr, queryMsg)
}

// ============================================================================
// SECURITY STATS
// ============================================================================

// GetSecurityStats returns security statistics
func (k Keeper) GetSecurityStats(ctx sdk.Context) types.SecurityStats {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.SecurityStatsKey)
	if bz == nil {
		return types.SecurityStats{}
	}

	var stats types.SecurityStats
	if err := json.Unmarshal(bz, &stats); err != nil {
		return types.SecurityStats{}
	}
	return stats
}

// SetSecurityStats sets security statistics (exported for testing)
func (k Keeper) SetSecurityStats(ctx sdk.Context, stats types.SecurityStats) {
	store := ctx.KVStore(k.storeKey)
	bz, err := json.Marshal(stats)
	if err != nil {
		return
	}
	store.Set(types.SecurityStatsKey, bz)
}

// setSecurityStats is an internal alias for SetSecurityStats
func (k Keeper) setSecurityStats(ctx sdk.Context, stats types.SecurityStats) {
	k.SetSecurityStats(ctx, stats)
}

// incrementSecurityStat increments a security statistic
func (k Keeper) incrementSecurityStat(ctx sdk.Context, statName string) {
	stats := k.GetSecurityStats(ctx)

	switch statName {
	case "contracts_uploaded":
		stats.TotalCodesAnalyzed++
	case "instantiated":
		// No specific field for instantiations, could use total_executions
		stats.TotalExecutions++
	case "executions":
		stats.TotalExecutions++
	case "contracts_migrated":
		// No field for migrations in SecurityStats
	case "paused":
		stats.ContractsPaused++
	case "reentrancy_blocked":
		// No field for reentrancy attempts, could use failed_executions
		stats.FailedExecutions++
	}

	k.setSecurityStats(ctx, stats)
}

// decrementSecurityStat decrements a security statistic
func (k Keeper) decrementSecurityStat(ctx sdk.Context, statName string) {
	stats := k.GetSecurityStats(ctx)

	switch statName {
	case "paused":
		if stats.ContractsPaused > 0 {
			stats.ContractsPaused--
		}
	}

	k.setSecurityStats(ctx, stats)
}

// ============================================================================
// EXECUTION CONTEXT (Reentrancy Protection)
// ============================================================================

// getOrCreateExecutionContext gets or creates execution context for this tx
func (k Keeper) getOrCreateExecutionContext(ctx sdk.Context) *types.ExecutionContext {
	// Use transient store keyed by tx hash
	tStore := ctx.TransientStore(k.storeKey)
	key := []byte("exec_ctx")

	bz := tStore.Get(key)
	if bz == nil {
		// Create new context
		execCtx := types.NewExecutionContext(10) // Max depth 10
		return execCtx
	}

	// Deserialize existing context
	var execCtx types.ExecutionContext
	if err := json.Unmarshal(bz, &execCtx); err != nil {
		// On error, create new context
		return types.NewExecutionContext(10)
	}

	return &execCtx
}

// setExecutionContext stores execution context
func (k Keeper) setExecutionContext(ctx sdk.Context, execCtx *types.ExecutionContext) {
	tStore := ctx.TransientStore(k.storeKey)
	key := []byte("exec_ctx")

	bz, err := json.Marshal(execCtx)
	if err != nil {
		k.Logger(ctx).Error("failed to marshal execution context", "error", err)
		return
	}

	tStore.Set(key, bz)
}

// ============================================================================
// SECURITY AUDIT LOGGING
// ============================================================================

// LogSecurityEvent logs a security audit event
func (k Keeper) LogSecurityEvent(ctx sdk.Context, event types.SecurityAuditEvent) {
	// Emit as SDK event
	attrs := []sdk.Attribute{
		sdk.NewAttribute("event_type", event.EventType),
		sdk.NewAttribute("contract", event.ContractAddr),
		sdk.NewAttribute("caller", event.Sender),
		sdk.NewAttribute("success", fmt.Sprintf("%t", event.Success)),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", event.BlockHeight)),
		sdk.NewAttribute("timestamp", event.Timestamp.String()),
	}

	if event.ErrorMessage != "" {
		attrs = append(attrs, sdk.NewAttribute("error", event.ErrorMessage))
	}

	if len(event.AdditionalData) > 0 {
		for k, v := range event.AdditionalData {
			attrs = append(attrs, sdk.NewAttribute(k, fmt.Sprintf("%v", v)))
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent("wasm_security_audit", attrs...),
	)
}
