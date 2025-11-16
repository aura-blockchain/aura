package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/aequitas/aura/chain/x/validatorsecurity/types"
)

type msgServer struct {
	types.UnimplementedMsgServer
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// RegisterValidator handles MsgRegisterValidator
func (ms msgServer) RegisterValidator(goCtx context.Context, msg *types.MsgRegisterValidator) (*types.MsgRegisterValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.RegisterValidator(
		ctx,
		msg.ValidatorAddress,
		msg.HotKey,
		msg.ColdKey,
		msg.Region,
		msg.CountryCode,
		msg.Latitude,
		msg.Longitude,
		msg.BackupValidatorAddresses,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_registered",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("region", msg.Region),
		),
	)

	return &types.MsgRegisterValidatorResponse{}, nil
}

// UpdateSecurityInfo handles MsgUpdateSecurityInfo
func (ms msgServer) UpdateSecurityInfo(goCtx context.Context, msg *types.MsgUpdateSecurityInfo) (*types.MsgUpdateSecurityInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	info, err := ms.Keeper.GetValidatorSecurityInfo(ctx, msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	// Update fields
	if msg.HotKey != "" {
		info.HotKey = msg.HotKey
	}
	if msg.ColdKey != "" {
		info.ColdKey = msg.ColdKey
	}
	if msg.Region != "" {
		// Update region count
		params := ms.Keeper.GetParams(ctx)
		if params.EnableGeoDistribution {
			if info.Region != msg.Region {
				ms.Keeper.decrementRegionCount(ctx, info.Region)
				if err := ms.Keeper.checkRegionCapacity(ctx, msg.Region); err != nil {
					return nil, err
				}
				ms.Keeper.incrementRegionCount(ctx, msg.Region)
			}
		}
		info.Region = msg.Region
	}
	if msg.CountryCode != "" {
		info.CountryCode = msg.CountryCode
	}
	if msg.Latitude != 0 {
		info.Latitude = msg.Latitude
	}
	if msg.Longitude != 0 {
		info.Longitude = msg.Longitude
	}
	if len(msg.BackupValidatorAddresses) > 0 {
		info.BackupValidatorAddresses = msg.BackupValidatorAddresses
	}

	info.KeysSeparated = info.HotKey != "" && info.ColdKey != "" && info.HotKey != info.ColdKey

	ms.Keeper.SetValidatorSecurityInfo(ctx, info)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_security_updated",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
		),
	)

	return &types.MsgUpdateSecurityInfoResponse{}, nil
}

// RegisterSentryNode handles MsgRegisterSentryNode
func (ms msgServer) RegisterSentryNode(goCtx context.Context, msg *types.MsgRegisterSentryNode) (*types.MsgRegisterSentryNodeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.RegisterSentryNode(
		ctx,
		msg.ValidatorAddress,
		msg.SentryAddress,
		msg.IpAddress,
		msg.Port,
	); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"sentry_node_registered",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("sentry", msg.SentryAddress),
		),
	)

	return &types.MsgRegisterSentryNodeResponse{}, nil
}

// ReportDoubleSign handles MsgReportDoubleSign
func (ms msgServer) ReportDoubleSign(goCtx context.Context, msg *types.MsgReportDoubleSign) (*types.MsgReportDoubleSignResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	slashAmount, err := ms.Keeper.HandleDoubleSign(
		ctx,
		msg.ValidatorAddress,
		msg.Height,
		msg.VoteA,
		msg.VoteB,
	)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"double_sign_reported",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
			sdk.NewAttribute("reporter", msg.ReporterAddress),
			sdk.NewAttribute("height", string(msg.Height)),
			sdk.NewAttribute("slash_amount", slashAmount.String()),
		),
	)

	return &types.MsgReportDoubleSignResponse{
		Slashed:     true,
		SlashAmount: slashAmount.String(),
	}, nil
}

// Unjail handles MsgUnjail
func (ms msgServer) Unjail(goCtx context.Context, msg *types.MsgUnjail) (*types.MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.UnjailValidator(ctx, msg.ValidatorAddress); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"validator_unjailed",
			sdk.NewAttribute("validator", msg.ValidatorAddress),
		),
	)

	return &types.MsgUnjailResponse{}, nil
}

// AcknowledgeAlert handles MsgAcknowledgeAlert
func (ms msgServer) AcknowledgeAlert(goCtx context.Context, msg *types.MsgAcknowledgeAlert) (*types.MsgAcknowledgeAlertResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := ms.Keeper.AcknowledgeAlert(ctx, msg.AlertId, msg.AcknowledgerAddress); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"alert_acknowledged",
			sdk.NewAttribute("alert_id", msg.AlertId),
			sdk.NewAttribute("acknowledger", msg.AcknowledgerAddress),
		),
	)

	return &types.MsgAcknowledgeAlertResponse{}, nil
}

// UpdateParams handles MsgUpdateParams
func (ms msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.Keeper.GetAuthority() != msg.Authority {
		return nil, types.ErrInvalidValidator
	}

	if err := ms.Keeper.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"params_updated",
			sdk.NewAttribute("authority", msg.Authority),
		),
	)

	return &types.MsgUpdateParamsResponse{}, nil
}
