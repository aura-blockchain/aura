package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/privacy/types"
	privacyproto "github.com/aequitas/aura/proto/aura/privacy/v1beta1"
)

// InitGenesisProto initializes the privacy module state from proto genesis
func (k *Keeper) InitGenesisProto(ctx context.Context, data *privacyproto.GenesisState) error {
	if data == nil {
		return nil
	}

	// Convert proto params to types.Params
	if data.Params != nil {
		params := types.Params{
			EnableZkProofs:                 data.Params.EnableZkProofs,
			EnableStealthAddresses:         data.Params.EnableStealthAddresses,
			EnableRingSignatures:           data.Params.EnableRingSignatures,
			EnableConfidentialTransactions: data.Params.EnableConfidentialTransactions,
			EnableNetworkPrivacy:           data.Params.EnableNetworkPrivacy,
			EnableMixing:                   data.Params.EnableMixing,
			MinRingSize:                    data.Params.MinRingSize,
			MaxRingSize:                    data.Params.MaxRingSize,
			MinMixingParticipants:          data.Params.MinMixingParticipants,
			MixingFee:                      data.Params.MixingFee,
			ZkProofVerificationCost:        data.Params.ZkProofVerificationCost,
		}

		// Validate params before setting
		if err := types.ValidateParams(params); err != nil {
			return err
		}

		if err := k.SetParams(ctx, params); err != nil {
			k.Logger(ctx).Error("failed to set params", "error", err)
			return err
		}
	}

	k.Logger(ctx).Info("privacy module initialized from genesis")
	return nil
}

// ExportGenesisProto exports the privacy module state to proto genesis
func (k *Keeper) ExportGenesisProto(ctx context.Context) *privacyproto.GenesisState {
	// Get parameters
	params := k.GetParams(ctx)

	protoParams := &privacyproto.Params{
		EnableZkProofs:                 params.EnableZkProofs,
		EnableStealthAddresses:         params.EnableStealthAddresses,
		EnableRingSignatures:           params.EnableRingSignatures,
		EnableConfidentialTransactions: params.EnableConfidentialTransactions,
		EnableNetworkPrivacy:           params.EnableNetworkPrivacy,
		EnableMixing:                   params.EnableMixing,
		MinRingSize:                    params.MinRingSize,
		MaxRingSize:                    params.MaxRingSize,
		MinMixingParticipants:          params.MinMixingParticipants,
		MixingFee:                      params.MixingFee,
		ZkProofVerificationCost:        params.ZkProofVerificationCost,
	}

	return &privacyproto.GenesisState{
		Params:             protoParams,
		MixingPools:        []*privacyproto.MixingPool{},
		RegisteredViewKeys: []*privacyproto.ViewKey{},
	}
}
