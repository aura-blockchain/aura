package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = msgServer{}

// msgServer is the gRPC server implementation for wasm module messages
type msgServer struct {
	types.UnimplementedMsgServer
	Keeper
}

// NewMsgServerImpl returns an implementation of the wasm MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// StoreCode handles uploading contract code
func (ms msgServer) StoreCode(goCtx context.Context, msg *types.MsgStoreCode) (*types.MsgStoreCodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	// Validate upload authorization and contract code
	if err := ms.Keeper.ValidateContractUpload(ctx, msg.Sender, msg.WASMByteCode); err != nil {
		return nil, err
	}

	// Store the code
	codeID, err := ms.Keeper.StoreCode(ctx, sender, msg.WASMByteCode)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeStoreCode,
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyCodeID, fmt.Sprintf("%d", codeID)),
		),
	})

	return &types.MsgStoreCodeResponse{
		CodeID: codeID,
	}, nil
}

// InstantiateContract handles contract instantiation
func (ms msgServer) InstantiateContract(goCtx context.Context, msg *types.MsgInstantiateContract) (*types.MsgInstantiateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	var admin sdk.AccAddress
	if msg.Admin != "" {
		admin, err = sdk.AccAddressFromBech32(msg.Admin)
		if err != nil {
			return nil, types.ErrInvalidAdmin.Wrapf("invalid admin address: %s", err)
		}
	}

	// Convert funds from proto type to sdk.Coins
	funds := sdk.NewCoins()
	for _, coin := range msg.Funds {
		if !coin.IsZero() {
			funds = funds.Add(coin)
		}
	}

	// Instantiate contract
	contractAddr, data, err := ms.Keeper.InstantiateContract(
		ctx,
		msg.CodeID,
		creator,
		admin,
		msg.Msg,
		msg.Label,
		funds,
	)
	if err != nil {
		return nil, err
	}

	// Store admin in AURA storage if admin was specified
	if !admin.Empty() {
		if err := ms.Keeper.SetContractAdmin(ctx, contractAddr, admin); err != nil {
			ms.Keeper.Logger(ctx).Error("failed to set contract admin in AURA storage",
				"contract", contractAddr.String(),
				"admin", admin.String(),
				"error", err)
			// Don't fail instantiation if admin storage fails, but log it
		}
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeInstantiate,
			sdk.NewAttribute(types.AttributeKeyContract, contractAddr.String()),
			sdk.NewAttribute(types.AttributeKeyCodeID, fmt.Sprintf("%d", msg.CodeID)),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgInstantiateContractResponse{
		Address: contractAddr.String(),
		Data:    data,
	}, nil
}

// ExecuteContract handles contract execution with enhanced reentrancy protection using call stack
func (ms msgServer) ExecuteContract(goCtx context.Context, msg *types.MsgExecuteContract) (*types.MsgExecuteContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Get or create execution context for this transaction
	execCtx := ms.Keeper.getOrCreateExecutionContext(ctx)

	// Try to push contract onto call stack (fails if reentrancy detected)
	if err := execCtx.PushContract(contractAddr.String()); err != nil {
		ms.Keeper.incrementSecurityStat(ctx, "reentrancy_blocked")

		// Log security event
		auditEvent := types.NewSecurityAuditEvent(
			"reentrancy_blocked",
			contractAddr.String(),
			sender.String(),
			ctx,
			false,
			err.Error(),
		)
		auditEvent.AddData("call_stack", execCtx.CallStack)
		auditEvent.AddData("call_depth", execCtx.CallDepth)
		ms.Keeper.LogSecurityEvent(ctx, auditEvent)

		return nil, err
	}

	// Ensure we pop the contract from stack on exit (cleanup)
	defer func() {
		if popErr := execCtx.PopContract(contractAddr.String()); popErr != nil {
			ms.Keeper.Logger(ctx).Error("failed to pop contract from call stack",
				"contract", contractAddr.String(),
				"error", popErr,
			)
		}
		// Save execution context back to transient store
		ms.Keeper.setExecutionContext(ctx, execCtx)
	}()

	// Record gas consumption start
	gasBefore := ctx.GasMeter().GasConsumed()

	// Convert funds from proto type to sdk.Coins
	funds := sdk.NewCoins()
	for _, coin := range msg.Funds {
		if !coin.IsZero() {
			funds = funds.Add(coin)
		}
	}

	// Execute contract with panic recovery
	var data []byte
	var execErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				execErr = types.ErrSecurityViolation.Wrapf("contract execution panicked: %v", r)
				ms.Keeper.Logger(ctx).Error("contract execution panic recovered",
					"contract", contractAddr.String(),
					"panic", r,
					"call_stack", execCtx.CallStack,
				)
			}
		}()

		data, execErr = ms.Keeper.ExecuteContract(ctx, contractAddr, sender, msg.Msg, funds)
	}()

	if execErr != nil {
		return nil, execErr
	}

	// Record gas consumed for this contract
	gasAfter := ctx.GasMeter().GasConsumed()
	execCtx.RecordGasConsumption(contractAddr.String(), gasAfter-gasBefore)

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeExecute,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
			sdk.NewAttribute("call_depth", fmt.Sprintf("%d", execCtx.CallDepth)),
		),
	})

	return &types.MsgExecuteContractResponse{
		Data: data,
	}, nil
}

// MigrateContract handles contract migration
func (ms msgServer) MigrateContract(goCtx context.Context, msg *types.MsgMigrateContract) (*types.MsgMigrateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	caller, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Verify caller is the contract admin
	isAdmin, err := ms.Keeper.IsContractAdmin(ctx, contractAddr, caller)
	if err != nil {
		return nil, types.ErrSecurityViolation.Wrapf("failed to verify admin: %s", err)
	}
	if !isAdmin {
		return nil, types.ErrUnauthorized.Wrapf("sender %s is not the contract admin", caller.String())
	}

	// Migrate contract
	data, err := ms.Keeper.Migrate(ctx, contractAddr, caller, msg.CodeID, msg.Msg)
	if err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMigrate,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			sdk.NewAttribute(types.AttributeKeyCodeID, fmt.Sprintf("%d", msg.CodeID)),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgMigrateContractResponse{
		Data: data,
	}, nil
}

// UpdateAdmin handles updating contract admin
func (ms msgServer) UpdateAdmin(goCtx context.Context, msg *types.MsgUpdateAdmin) (*types.MsgUpdateAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	newAdmin, err := sdk.AccAddressFromBech32(msg.NewAdmin)
	if err != nil {
		return nil, types.ErrInvalidAdmin.Wrapf("invalid new admin address: %s", err)
	}

	// Verify sender is the current admin
	currentAdmin, err := ms.Keeper.GetContractAdmin(ctx, contractAddr)
	if err != nil {
		return nil, types.ErrSecurityViolation.Wrapf("failed to get current admin: %s", err)
	}

	// If no admin is set in AURA storage, this is an unauthorized operation
	// (contracts must be instantiated through AURA's InstantiateContract which sets admin)
	if currentAdmin.Empty() {
		return nil, types.ErrUnauthorized.Wrap("contract has no admin set")
	}

	// Verify sender is the current admin
	if !currentAdmin.Equals(sender) {
		return nil, types.ErrUnauthorized.Wrapf("sender %s is not the contract admin %s",
			sender.String(), currentAdmin.String())
	}

	// Update admin in AURA storage
	if err := ms.Keeper.SetContractAdmin(ctx, contractAddr, newAdmin); err != nil {
		return nil, types.ErrSecurityViolation.Wrapf("failed to update admin: %s", err)
	}

	// Also update in wasmd keeper if available (dual storage for compatibility)
	if ms.Keeper.wasmKeeper != nil {
		ops := wasmkeeper.NewDefaultPermissionKeeper(ms.Keeper.wasmKeeper)
		if err := ops.UpdateContractAdmin(ctx, contractAddr, sender, newAdmin); err != nil {
			ms.Keeper.Logger(ctx).Error("failed to update admin in wasmd keeper",
				"contract", msg.Contract,
				"error", err)
			// Don't fail the transaction - AURA storage is source of truth
		}
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateAdmin,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			sdk.NewAttribute(types.AttributeKeyNewAdmin, msg.NewAdmin),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgUpdateAdminResponse{}, nil
}

// ClearAdmin handles clearing contract admin
func (ms msgServer) ClearAdmin(goCtx context.Context, msg *types.MsgClearAdmin) (*types.MsgClearAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Verify sender is the current admin
	currentAdmin, err := ms.Keeper.GetContractAdmin(ctx, contractAddr)
	if err != nil {
		return nil, types.ErrSecurityViolation.Wrapf("failed to get current admin: %s", err)
	}

	// If no admin is set, this is an error
	if currentAdmin.Empty() {
		return nil, types.ErrUnauthorized.Wrap("contract has no admin to clear")
	}

	// Verify sender is the current admin
	if !currentAdmin.Equals(sender) {
		return nil, types.ErrUnauthorized.Wrapf("sender %s is not the contract admin %s",
			sender.String(), currentAdmin.String())
	}

	// Clear admin in AURA storage
	if err := ms.Keeper.DeleteContractAdmin(ctx, contractAddr); err != nil {
		return nil, types.ErrSecurityViolation.Wrapf("failed to clear admin: %s", err)
	}

	// Also clear in wasmd keeper if available (dual storage for compatibility)
	if ms.Keeper.wasmKeeper != nil {
		ops := wasmkeeper.NewDefaultPermissionKeeper(ms.Keeper.wasmKeeper)
		if err := ops.ClearContractAdmin(ctx, contractAddr, sender); err != nil {
			ms.Keeper.Logger(ctx).Error("failed to clear admin in wasmd keeper",
				"contract", msg.Contract,
				"error", err)
			// Don't fail the transaction - AURA storage is source of truth
		}
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClearAdmin,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			sdk.NewAttribute(types.AttributeKeySender, msg.Sender),
		),
	})

	return &types.MsgClearAdminResponse{}, nil
}

// AuthorizeUploader handles authorizing contract uploaders (governance only)
func (ms msgServer) AuthorizeUploader(goCtx context.Context, msg *types.MsgAuthorizeUploader) (*types.MsgAuthorizeUploaderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != ms.Keeper.authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.Keeper.authority, msg.Authority)
	}

	// Validate uploader address
	if _, err := sdk.AccAddressFromBech32(msg.Uploader); err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid uploader address: %s", err)
	}

	// Authorize uploader
	if err := ms.Keeper.AuthorizeUploader(ctx, msg.Uploader); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAuthorizeUploader,
			sdk.NewAttribute(types.AttributeKeyUploader, msg.Uploader),
		),
	})

	return &types.MsgAuthorizeUploaderResponse{}, nil
}

// RevokeUploader handles revoking contract uploader authorization (governance only)
func (ms msgServer) RevokeUploader(goCtx context.Context, msg *types.MsgRevokeUploader) (*types.MsgRevokeUploaderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != ms.Keeper.authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.Keeper.authority, msg.Authority)
	}

	// Validate uploader address
	if _, err := sdk.AccAddressFromBech32(msg.Uploader); err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid uploader address: %s", err)
	}

	// Revoke uploader
	if err := ms.Keeper.RevokeUploader(ctx, msg.Uploader); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRevokeUploader,
			sdk.NewAttribute(types.AttributeKeyUploader, msg.Uploader),
		),
	})

	return &types.MsgRevokeUploaderResponse{}, nil
}

// PauseContract handles pausing a contract (governance only)
func (ms msgServer) PauseContract(goCtx context.Context, msg *types.MsgPauseContract) (*types.MsgPauseContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != ms.Keeper.authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.Keeper.authority, msg.Authority)
	}

	// Validate contract address
	if _, err := sdk.AccAddressFromBech32(msg.Contract); err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Pause contract
	if err := ms.Keeper.PauseContract(ctx, msg.Contract); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypePauseContract,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
		),
	})

	return &types.MsgPauseContractResponse{}, nil
}

// UnpauseContract handles unpausing a contract (governance only)
func (ms msgServer) UnpauseContract(goCtx context.Context, msg *types.MsgUnpauseContract) (*types.MsgUnpauseContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check authority
	if msg.Authority != ms.Keeper.authority {
		return nil, types.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s", ms.Keeper.authority, msg.Authority)
	}

	// Validate contract address
	if _, err := sdk.AccAddressFromBech32(msg.Contract); err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Unpause contract
	if err := ms.Keeper.UnpauseContract(ctx, msg.Contract); err != nil {
		return nil, err
	}

	// Emit event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnpauseContract,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
		),
	})

	return &types.MsgUnpauseContractResponse{}, nil
}

// UpdateParams handles updating module parameters (governance only)
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update params using keeper method which validates authority
	if err := ms.Keeper.UpdateParams(goCtx, msg.Authority, msg.Params); err != nil {
		return nil, err
	}

	// Emit event
	paramsJSON, _ := json.Marshal(msg.Params)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateParams,
			sdk.NewAttribute(types.AttributeKeyParams, string(paramsJSON)),
		),
	})

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetExecuting marks a contract as executing or not (exported for testing)
func (k Keeper) SetExecuting(ctx sdk.Context, contractAddr string, executing bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractExecutingKey(contractAddr)
	if executing {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
}
