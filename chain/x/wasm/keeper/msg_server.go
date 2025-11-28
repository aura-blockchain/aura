package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aequitas/aura/chain/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.MsgServer = msgServer{}

// msgServer is the gRPC server implementation for wasm module messages
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the wasm MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// StoreCode handles uploading contract code
func (ms msgServer) StoreCode(goCtx context.Context, msg *types.MsgStoreCode) (*types.MsgStoreCodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Instantiate contract
	contractAddr, data, err := ms.Keeper.InstantiateContract(
		ctx,
		msg.CodeID,
		creator,
		admin,
		msg.Msg,
		msg.Label,
		msg.Funds,
	)
	if err != nil {
		return nil, err
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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

		data, execErr = ms.Keeper.ExecuteContract(ctx, contractAddr, sender, msg.Msg, msg.Funds)
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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	caller, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	contractAddr, err := sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.NewAdmin)
	if err != nil {
		return nil, types.ErrInvalidAdmin.Wrapf("invalid new admin address: %s", err)
	}

	// Note: In production, this would call wasmd keeper to update admin
	// For now, we just emit the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUpdateAdmin,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
			sdk.NewAttribute(types.AttributeKeyNewAdmin, msg.NewAdmin),
		),
	})

	return &types.MsgUpdateAdminResponse{}, nil
}

// ClearAdmin handles clearing contract admin
func (ms msgServer) ClearAdmin(goCtx context.Context, msg *types.MsgClearAdmin) (*types.MsgClearAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, types.ErrUnauthorized.Wrapf("invalid sender address: %s", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Contract)
	if err != nil {
		return nil, types.ErrInvalidContractAddress.Wrapf("invalid contract address: %s", err)
	}

	// Note: In production, this would call wasmd keeper to clear admin
	// For now, we just emit the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClearAdmin,
			sdk.NewAttribute(types.AttributeKeyContract, msg.Contract),
		),
	})

	return &types.MsgClearAdminResponse{}, nil
}

// AuthorizeUploader handles authorizing contract uploaders (governance only)
func (ms msgServer) AuthorizeUploader(goCtx context.Context, msg *types.MsgAuthorizeUploader) (*types.MsgAuthorizeUploaderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

	// Validate message
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

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

// isReentrancyAttempt checks if a contract is currently executing
func (k Keeper) isReentrancyAttempt(ctx sdk.Context, contractAddr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetContractExecutingKey(contractAddr)
	return store.Has(key)
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

// setExecuting is an internal alias for SetExecuting
func (k Keeper) setExecuting(ctx sdk.Context, contractAddr string, executing bool) {
	k.SetExecuting(ctx, contractAddr, executing)
}
