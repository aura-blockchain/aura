package keeper

import (
	"fmt"

	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.MsgServer = msgServer{}

type msgServer struct {
	pb.UnimplementedMsgServer
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) pb.MsgServer {
	return &msgServer{Keeper: keeper}
}

// RegisterContract handles contract registration
func (ms msgServer) RegisterContract(goCtx context.Context, msg *pb.MsgRegisterContract) (*pb.MsgRegisterContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Authorization check - signer must be creator or authority
	if msg.Signer != msg.Creator && msg.Signer != ms.GetAuthority() {
		return nil, mapAuthzError(types.ErrUnauthorized)
	}

	// Check max contracts per creator limit
	params := ms.Keeper.GetParams(ctx)
	if params.MaxContractsPerCreator > 0 {
		creatorContracts := ms.Keeper.GetCreatorContracts(ctx, msg.Creator)
		if uint64(len(creatorContracts)) >= params.MaxContractsPerCreator {
			return nil, types.ErrTooManyContracts
		}
	}

	// Build contract info from proto message
	info := &pb.ContractInfo{
		Address:        msg.ContractAddress,
		CodeId:         msg.CodeId,
		Creator:        msg.Creator,
		Admin:          msg.Admin,
		Label:          msg.Label,
		Status:         pb.ContractStatus_CONTRACT_STATUS_ACTIVE,
		Metadata:       msg.Metadata,
		SecurityPolicy: msg.SecurityPolicy,
		Compliance:     msg.Compliance,
	}

	// Register contract
	if err := ms.Keeper.RegisterContract(ctx, info); err != nil {
		return nil, err
	}

	// Add to creator index
	ms.Keeper.AddCreatorContract(ctx, msg.Creator, msg.ContractAddress)

	// Add to tag indices if metadata has tags
	// Metadata is a value type, always present
	if len(msg.Metadata.Tags) > 0 {
		for _, tag := range msg.Metadata.Tags {
			ms.Keeper.AddTagContract(ctx, tag, msg.ContractAddress)
		}
	}

	return &pb.MsgRegisterContractResponse{
		Success:         true,
		ContractAddress: msg.ContractAddress,
	}, nil
}

// UpdateContractMetadata handles metadata updates
func (ms msgServer) UpdateContractMetadata(goCtx context.Context, msg *pb.MsgUpdateContractMetadata) (*pb.MsgUpdateContractMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update metadata - pass pointer to value type
	if err := ms.Keeper.UpdateContractMetadata(ctx, msg.ContractAddress, msg.Signer, &msg.Metadata); err != nil {
		return nil, mapAuthzError(err)
	}

	return &pb.MsgUpdateContractMetadataResponse{
		Success: true,
	}, nil
}

// UpdateSecurityPolicy handles security policy updates
func (ms msgServer) UpdateSecurityPolicy(goCtx context.Context, msg *pb.MsgUpdateSecurityPolicy) (*pb.MsgUpdateSecurityPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Update security policy - pass pointer to value type
	if err := ms.Keeper.UpdateSecurityPolicy(ctx, msg.ContractAddress, msg.Signer, &msg.SecurityPolicy); err != nil {
		return nil, mapAuthzError(err)
	}

	return &pb.MsgUpdateSecurityPolicyResponse{
		Success: true,
	}, nil
}

// PauseContract handles contract pausing
func (ms msgServer) PauseContract(goCtx context.Context, msg *pb.MsgPauseContract) (*pb.MsgPauseContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract exists
	info, found := ms.Keeper.GetContractInfo(ctx, msg.ContractAddress)
	if !found {
		return nil, types.ErrContractNotFound
	}

	// Check if pause is allowed by security policy
	// SecurityPolicy is a value type, always present
	if !info.SecurityPolicy.AllowPause {
		return nil, mapAuthzError(types.ErrUnauthorized)
	}

	// Verify authorization - must be admin or authority
	if info.Admin != msg.Signer && msg.Signer != ms.GetAuthority() {
		return nil, mapAuthzError(types.ErrUnauthorized)
	}

	// Pause contract
	if err := ms.Keeper.PauseContract(ctx, msg.ContractAddress, msg.Signer, msg.Reason); err != nil {
		return nil, mapAuthzError(err)
	}

	return &pb.MsgPauseContractResponse{
		Success: true,
	}, nil
}

// UnpauseContract handles contract unpausing
func (ms msgServer) UnpauseContract(goCtx context.Context, msg *pb.MsgUnpauseContract) (*pb.MsgUnpauseContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract exists and is paused
	info, found := ms.Keeper.GetContractInfo(ctx, msg.ContractAddress)
	if !found {
		return nil, types.ErrContractNotFound
	}

	// Verify contract is actually paused
	if info.Status != pb.ContractStatus_CONTRACT_STATUS_PAUSED {
		return nil, types.ErrContractNotActive
	}

	// Unpause contract
	if err := ms.Keeper.UnpauseContract(ctx, msg.ContractAddress, msg.Signer); err != nil {
		return nil, mapAuthzError(err)
	}

	return &pb.MsgUnpauseContractResponse{
		Success: true,
	}, nil
}

// DeprecateContract handles contract deprecation
func (ms msgServer) DeprecateContract(goCtx context.Context, msg *pb.MsgDeprecateContract) (*pb.MsgDeprecateContractResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if contract exists
	info, found := ms.Keeper.GetContractInfo(ctx, msg.ContractAddress)
	if !found {
		return nil, types.ErrContractNotFound
	}

	// Verify authorization - must be admin or authority
	if info.Admin != msg.Signer && msg.Signer != ms.GetAuthority() {
		return nil, mapAuthzError(types.ErrUnauthorized)
	}

	// Deprecate contract
	if err := ms.Keeper.DeprecateContract(ctx, msg.ContractAddress, msg.Signer, msg.Reason, msg.MigrationTarget); err != nil {
		return nil, mapAuthzError(err)
	}

	return &pb.MsgDeprecateContractResponse{
		Success: true,
	}, nil
}

func mapAuthzError(err error) error {
	if err == nil {
		return nil
	}

	if errorsmod.IsOf(err, types.ErrUnauthorized) ||
		errorsmod.IsOf(err, types.ErrNotContractAdmin) ||
		errorsmod.IsOf(err, types.ErrNotContractCreator) ||
		errorsmod.IsOf(err, types.ErrInvalidSigner) {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	return fmt.Errorf("error in mapAuthzError for ErrInvalidSigner: %w", err)
}
