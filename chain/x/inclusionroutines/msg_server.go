package inclusionroutines

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/inclusionroutines/keeper"
	"github.com/aequitas/aura/chain/x/inclusionroutines/types"
	inclusionroutinespb "github.com/aequitas/aura/proto/aura/inclusionroutines/v1beta1"
)

type msgServer struct {
	inclusionroutinespb.UnimplementedMsgServer
	keeper *keeper.Keeper
}

// NewMsgServer creates a new message server
func NewMsgServer(k *keeper.Keeper) inclusionroutinespb.MsgServer {
	return &msgServer{keeper: k}
}

// CreateIR handles the creation of a new IR definition
func (s *msgServer) CreateIR(ctx context.Context, msg *inclusionroutinespb.MsgCreateIR) (*inclusionroutinespb.MsgCreateIRResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	ir := types.IRDefinition{
		ID:               msg.Id,
		Name:             msg.Name,
		Arena:            msg.Arena,
		Description:      msg.Description,
		Score:            msg.Score,
		POIReward:        msg.PoiReward,
		LocaleTags:       msg.LocaleTags,
		PrivacyTier:      msg.PrivacyTier,
		Version:          msg.Version,
		MetadataHash:     msg.MetadataHash,
		Status:           inclusionroutinespb.IRStatus_IR_STATUS_DRAFT,
		ActivationHeight: msg.ActivationHeight,
		SunsetHeight:     msg.SunsetHeight,
	}

	if err := s.keeper.CreateIR(ir); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgCreateIRResponse{Id: msg.Id}, nil
}

// UpdateIR handles updates to an existing IR definition
func (s *msgServer) UpdateIR(ctx context.Context, msg *inclusionroutinespb.MsgUpdateIR) (*inclusionroutinespb.MsgUpdateIRResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	// Get existing IR
	existing, ok := s.keeper.GetIR(msg.Id)
	if !ok {
		return nil, types.ErrIRNotFound
	}

	// Update mutable fields
	if msg.Name != "" {
		existing.Name = msg.Name
	}
	if msg.Description != "" {
		existing.Description = msg.Description
	}
	if msg.Score > 0 {
		existing.Score = msg.Score
	}
	if msg.PoiReward > 0 {
		existing.POIReward = msg.PoiReward
	}
	if len(msg.LocaleTags) > 0 {
		existing.LocaleTags = msg.LocaleTags
	}
	if msg.PrivacyTier != inclusionroutinespb.PrivacyTier_PRIVACY_TIER_UNSPECIFIED {
		existing.PrivacyTier = msg.PrivacyTier
	}
	if msg.Version != "" {
		existing.Version = msg.Version
	}
	if msg.MetadataHash != "" {
		existing.MetadataHash = msg.MetadataHash
	}
	if msg.SunsetHeight > 0 {
		existing.SunsetHeight = msg.SunsetHeight
	}

	if err := s.keeper.UpdateIR(existing); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgUpdateIRResponse{}, nil
}

// DeleteIR handles deletion of an IR definition
func (s *msgServer) DeleteIR(ctx context.Context, msg *inclusionroutinespb.MsgDeleteIR) (*inclusionroutinespb.MsgDeleteIRResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	if err := s.keeper.DeleteIR(msg.Id); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgDeleteIRResponse{}, nil
}

// SetIRPrerequisites handles setting prerequisites for an IR
func (s *msgServer) SetIRPrerequisites(ctx context.Context, msg *inclusionroutinespb.MsgSetIRPrerequisites) (*inclusionroutinespb.MsgSetIRPrerequisitesResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	if err := s.keeper.SetPrerequisites(msg.IrId, msg.RequiredIrIds); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgSetIRPrerequisitesResponse{}, nil
}

// SetIRRateLimit handles setting rate limits for an IR
func (s *msgServer) SetIRRateLimit(ctx context.Context, msg *inclusionroutinespb.MsgSetIRRateLimit) (*inclusionroutinespb.MsgSetIRRateLimitResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	limit := types.IRRateLimit{
		IRID:             msg.IrId,
		PerWalletPerHour: msg.PerWalletPerHour,
		PerWalletPerDay:  msg.PerWalletPerDay,
		PerBlockGlobal:   msg.PerBlockGlobal,
	}

	if err := s.keeper.SetRateLimit(limit); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgSetIRRateLimitResponse{}, nil
}

// SuspendIR handles suspending an IR
func (s *msgServer) SuspendIR(ctx context.Context, msg *inclusionroutinespb.MsgSuspendIR) (*inclusionroutinespb.MsgSuspendIRResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	if err := s.keeper.SuspendIR(msg.IrId); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgSuspendIRResponse{}, nil
}

// ActivateIR handles activating an IR
func (s *msgServer) ActivateIR(ctx context.Context, msg *inclusionroutinespb.MsgActivateIR) (*inclusionroutinespb.MsgActivateIRResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("msg required")
	}

	// Verify authority is governance
	if msg.Authority != s.keeper.GetAuthority() {
		return nil, fmt.Errorf("invalid authority; expected %s, got %s", s.keeper.GetAuthority(), msg.Authority)
	}

	if err := s.keeper.ActivateIR(msg.IrId); err != nil {
		return nil, err
	}

	return &inclusionroutinespb.MsgActivateIRResponse{}, nil
}
