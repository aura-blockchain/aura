// Copyright 2024-2025 Aequitas Foundation
// SPDX-License-Identifier: Apache-2.0

package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aequitas/aura/chain/x/dataregistry/types"
	pb "github.com/aequitas/aura/proto/aura/dataregistry/v1beta1"
)

// Ensure msgServer implements MsgServer interface
var _ pb.MsgServer = &msgServer{}

// msgServer implements the gRPC MsgServer interface
type msgServer struct {
	pb.UnimplementedMsgServer
	keeper *Keeper
}

// NewMsgServer creates a new MsgServer instance
func NewMsgServer(k *Keeper) pb.MsgServer {
	return &msgServer{keeper: k}
}

// StoreDataItem stores a new data item
func (s *msgServer) StoreDataItem(ctx context.Context, msg *pb.MsgStoreDataItem) (*pb.MsgStoreDataItemResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Validate message
	if err := s.validateMsgStoreDataItem(msg); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Store the data item
	dataID, err := s.keeper.StoreDataItem(
		sdkCtx,
		msg.Creator,
		msg.DataType,
		msg.Title,
		msg.Description,
		msg.ContentHash,
		msg.StorageLocation,
		msg.IsEncrypted,
		msg.GeoLocation,
		msg.Metadata,
		msg.AccessPolicy,
		msg.Tags,
	)
	if err != nil {
		return nil, err
	}

	// Get the stored item to retrieve the creation timestamp
	item, exists := s.keeper.GetDataItem(sdkCtx, dataID)
	if !exists {
		return nil, fmt.Errorf("data item was stored but could not be retrieved")
	}

	if sdkCtx, ok := sdkContextFromCtx(ctx); ok {
		emitDataItemEvent(
			sdkCtx,
			types.EventTypeDataItemStored,
			sdk.NewAttribute(types.AttributeKeyDataID, dataID),
			sdk.NewAttribute(types.AttributeKeyDataType, fmt.Sprintf("%d", msg.DataType)),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyStorageLocation, msg.StorageLocation),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
		)
	}

	return &pb.MsgStoreDataItemResponse{
		DataId:    dataID,
		CreatedAt: item.CreatedAt,
	}, nil
}

// UpdateDataItem updates an existing data item
func (s *msgServer) UpdateDataItem(ctx context.Context, msg *pb.MsgUpdateDataItem) (*pb.MsgUpdateDataItemResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Validate message
	if err := s.validateMsgUpdateDataItem(msg); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Update the data item
	err := s.keeper.UpdateDataItem(
		sdkCtx,
		msg.DataId,
		msg.Creator,
		msg.Title,
		msg.Description,
		msg.Metadata,
		msg.AccessPolicy,
		msg.Tags,
	)
	if err != nil {
		return nil, err
	}

	updatedAt := timestampFromTime(sdk.UnwrapSDKContext(ctx).BlockTime())

	if sdkCtx, ok := sdkContextFromCtx(ctx); ok {
		sdkCtx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeDataItemUpdated,
				sdk.NewAttribute(types.AttributeKeyDataID, msg.DataId),
				sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
				sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
				sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
			),
		})
		emitDataItemEvent(
			sdkCtx,
			types.EventTypeDataItemUpdated,
			sdk.NewAttribute(types.AttributeKeyDataID, msg.DataId),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
		)
	}

	return &pb.MsgUpdateDataItemResponse{
		UpdatedAt: updatedAt,
	}, nil
}

// DeleteDataItem deletes a data item
func (s *msgServer) DeleteDataItem(ctx context.Context, msg *pb.MsgDeleteDataItem) (*pb.MsgDeleteDataItemResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Validate message
	if err := s.validateMsgDeleteDataItem(msg); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get item to verify ownership
	item, ok := s.keeper.GetDataItem(sdkCtx, msg.DataId)
	if !ok {
		return nil, types.ErrDataItemNotFound
	}

	if item.OwnerAddress != msg.Creator {
		return nil, status.Error(codes.PermissionDenied, types.ErrUnauthorized.Error())
	}

	// Delete the data item
	err := s.keeper.DeleteDataItem(sdkCtx, msg.DataId)
	if err != nil {
		return nil, err
	}

	deletedAt := timestampFromTime(sdk.UnwrapSDKContext(ctx).BlockTime())

	if sdkCtx, ok := sdkContextFromCtx(ctx); ok {
		emitDataItemEvent(
			sdkCtx,
			types.EventTypeDataItemDeleted,
			sdk.NewAttribute(types.AttributeKeyDataID, msg.DataId),
			sdk.NewAttribute(types.AttributeKeyOwner, msg.Creator),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
		)
	}

	return &pb.MsgDeleteDataItemResponse{
		DeletedAt: deletedAt,
	}, nil
}

// VerifyDataItem adds verification to a data item
func (s *msgServer) VerifyDataItem(ctx context.Context, msg *pb.MsgVerifyDataItem) (*pb.MsgVerifyDataItemResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Validate message
	if err := s.validateMsgVerifyDataItem(msg); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Verify the data item
	err := s.keeper.VerifyDataItem(
		sdkCtx,
		msg.DataId,
		msg.Verifier,
		msg.Level,
		msg.ConfidenceScore,
		msg.Notes,
		msg.VerificationMethod,
		msg.Proof,
	)
	if err != nil {
		return nil, err
	}

	// Get verification reward from params
	params, _ := s.keeper.GetParams(ctx)

	verifiedAt := timestampFromTime(sdk.UnwrapSDKContext(ctx).BlockTime())

	if sdkCtx, ok := sdkContextFromCtx(ctx); ok {
		emitDataItemEvent(
			sdkCtx,
			types.EventTypeDataItemVerified,
			sdk.NewAttribute(types.AttributeKeyDataID, msg.DataId),
			sdk.NewAttribute(types.AttributeKeyVerifier, msg.Verifier),
			sdk.NewAttribute(types.AttributeKeyVerificationLvl, fmt.Sprintf("%d", msg.Level)),
			sdk.NewAttribute(types.AttributeKeyConfidenceScore, fmt.Sprintf("%d", msg.ConfidenceScore)),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
		)
	}

	if params.VerificationReward > 0 {
		verifierAddr, err := sdk.AccAddressFromBech32(msg.Verifier)
		if err == nil {
			rewardCoins := sdk.NewCoins(sdk.NewInt64Coin("uaura", int64(params.VerificationReward)))

			if s.keeper.bankKeeper != nil {
				if err := s.keeper.bankKeeper.MintCoins(sdkCtx, types.ModuleName, rewardCoins); err == nil {
					if err := s.keeper.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, verifierAddr, rewardCoins); err == nil {
						s.keeper.Logger(sdkCtx).Info("minted verification reward",
							"verifier", msg.Verifier,
							"amount", params.VerificationReward)
					}
				}
			}
		}
	}

	return &pb.MsgVerifyDataItemResponse{
		VerifiedAt:         verifiedAt,
		VerificationReward: params.VerificationReward,
	}, nil
}

// RevokeDataItem revokes a data item
func (s *msgServer) RevokeDataItem(ctx context.Context, msg *pb.MsgRevokeDataItem) (*pb.MsgRevokeDataItemResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("message cannot be nil")
	}

	// Validate message
	if err := s.validateMsgRevokeDataItem(msg); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, _ := s.keeper.GetParams(ctx)
	isAuthorized := false

	for _, authorizedVerifier := range params.AuthorizedVerifiers {
		if authorizedVerifier == msg.Authority {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		if s.keeper.authority != "" && msg.Authority == s.keeper.authority {
			isAuthorized = true
		}
	}

	if !isAuthorized {
		return nil, fmt.Errorf("authority %s is not authorized to revoke data items", msg.Authority)
	}

	// Revoke the data item
	err := s.keeper.RevokeDataItem(
		sdkCtx,
		msg.DataId,
		msg.Authority,
		msg.Reason,
	)
	if err != nil {
		return nil, err
	}

	revokedAt := timestampFromTime(sdk.UnwrapSDKContext(ctx).BlockTime())

	if sdkCtx, ok := sdkContextFromCtx(ctx); ok {
		emitDataItemEvent(
			sdkCtx,
			types.EventTypeDataItemRevoked,
			sdk.NewAttribute(types.AttributeKeyDataID, msg.DataId),
			sdk.NewAttribute(types.AttributeKeyAuthority, msg.Authority),
			sdk.NewAttribute(types.AttributeKeyReason, msg.Reason),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
			sdk.NewAttribute(types.AttributeKeyTimestamp, sdkCtx.BlockTime().UTC().Format(time.RFC3339Nano)),
		)
	}

	return &pb.MsgRevokeDataItemResponse{
		RevokedAt: revokedAt,
	}, nil
}

// ============================
// VALIDATION HELPERS
// ============================

func (s *msgServer) validateMsgStoreDataItem(msg *pb.MsgStoreDataItem) error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}

	if msg.DataType == pb.DataItemType_DATA_ITEM_TYPE_UNSPECIFIED {
		return types.ErrInvalidDataType
	}

	if msg.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if len(msg.ContentHash) == 0 {
		return types.ErrInvalidContentHash
	}

	if msg.StorageLocation == "" {
		return types.ErrInvalidStorageLocation
	}

	// Validate access policy if provided
	if msg.AccessPolicy != nil {
		if err := validateAccessPolicy(msg.AccessPolicy); err != nil {
			return fmt.Errorf("error in validateMsgStoreDataItem for ErrInvalidContentHash: %w", err)
		}
	}

	return nil
}

func (s *msgServer) validateMsgUpdateDataItem(msg *pb.MsgUpdateDataItem) error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}

	if msg.DataId == "" {
		return fmt.Errorf("data_id cannot be empty")
	}

	// At least one field must be updated
	if msg.Title == "" && msg.Description == "" && len(msg.Metadata) == 0 && msg.AccessPolicy == nil && len(msg.Tags) == 0 {
		return fmt.Errorf("at least one field must be provided for update")
	}

	// Validate access policy if provided
	if msg.AccessPolicy != nil {
		if err := validateAccessPolicy(msg.AccessPolicy); err != nil {
			return fmt.Errorf("error in validateMsgUpdateDataItem for provided: %w", err)
		}
	}

	return nil
}

func (s *msgServer) validateMsgDeleteDataItem(msg *pb.MsgDeleteDataItem) error {
	if msg.Creator == "" {
		return fmt.Errorf("creator cannot be empty")
	}

	if msg.DataId == "" {
		return fmt.Errorf("data_id cannot be empty")
	}

	return nil
}

func (s *msgServer) validateMsgVerifyDataItem(msg *pb.MsgVerifyDataItem) error {
	if msg.Verifier == "" {
		return fmt.Errorf("verifier cannot be empty")
	}

	if msg.DataId == "" {
		return fmt.Errorf("data_id cannot be empty")
	}

	if msg.Level == pb.VerificationLevel_VERIFICATION_LEVEL_UNSPECIFIED {
		return types.ErrInvalidVerificationLevel
	}

	if msg.ConfidenceScore > 100 {
		return fmt.Errorf("confidence_score must be between 0 and 100")
	}

	if msg.VerificationMethod == "" {
		return fmt.Errorf("verification_method cannot be empty")
	}

	return nil
}

func (s *msgServer) validateMsgRevokeDataItem(msg *pb.MsgRevokeDataItem) error {
	if msg.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}

	if msg.DataId == "" {
		return fmt.Errorf("data_id cannot be empty")
	}

	if msg.Reason == "" {
		return fmt.Errorf("reason cannot be empty")
	}

	return nil
}

func validateAccessPolicy(policy *pb.AccessPolicy) error {
	if policy == nil {
		return nil
	}

	// Validate mode-specific rules
	switch policy.Mode {
	case pb.AccessMode_ACCESS_MODE_WHITELIST:
		if len(policy.AllowedAddresses) == 0 {
			return fmt.Errorf("whitelist mode requires at least one allowed address")
		}
	case pb.AccessMode_ACCESS_MODE_PRIVATE:
		// Private mode should not have allowed addresses
		if len(policy.AllowedAddresses) > 0 {
			return fmt.Errorf("private mode should not have allowed addresses")
		}
	}

	return nil
}

func emitDataItemEvent(ctx sdk.Context, eventType string, attrs ...sdk.Attribute) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(eventType, attrs...),
	)
}

func sdkContextFromCtx(ctx context.Context) (sdk.Context, bool) {
	var sdkCtx sdk.Context
	ok := true

	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()

	sdkCtx = sdk.UnwrapSDKContext(ctx)
	return sdkCtx, ok
}
