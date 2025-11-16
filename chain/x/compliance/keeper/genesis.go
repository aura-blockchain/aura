package keeper

import (
	"context"

	"github.com/aequitas/aura/chain/x/compliance/types"
	complianceproto "github.com/aequitas/aura/proto/aura/compliance/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the compliance module state from genesis
func (k *Keeper) InitGenesis(ctx context.Context, data *complianceproto.GenesisState) error {
	if data == nil {
		return nil
	}

	// Convert context to SDK context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Set parameters - convert proto params to local types
	params := types.ComplianceParams{
		KYCRequired:                  data.Params.KycRequired,
		MinimumKYCLevel:              types.KYCLevel(data.Params.MinimumKycLevel),
		KYCExpiryDays:                data.Params.KycExpiryDays,
		TransactionMonitoringEnabled: data.Params.TransactionMonitoringEnabled,
		VelocityLimit24h:             data.Params.VelocityLimit_24H,
		SingleTransactionLimit:       data.Params.SingleTransactionLimit,
		StructuringThresholdCount:    data.Params.StructuringThresholdCount,
		SanctionsScreeningEnabled:    data.Params.SanctionsScreeningEnabled,
		SanctionsList:                data.Params.SanctionsList,
		ScreeningCacheHours:          data.Params.ScreeningCacheHours,
		GDPREnabled:                  data.Params.GdprEnabled,
		DataRetentionDays:            data.Params.DataRetentionDays,
		ProcessingPurposes:           data.Params.ProcessingPurposes,
		TaxReportingEnabled:          data.Params.TaxReportingEnabled,
		TaxJurisdictions:             data.Params.TaxJurisdictions,
		TaxYearEnd:                   data.Params.TaxYearEnd,
	}

	if err := k.SetParams(sdkCtx, params); err != nil {
		return err
	}

	// TODO: Initialize other genesis data once keeper methods are properly implemented with KVStore
	// For now, just initialize parameters to allow chain to start

	return nil
}

// ExportGenesis exports the compliance module state to genesis
func (k *Keeper) ExportGenesis(ctx context.Context) *complianceproto.GenesisState {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	// Convert local params to proto
	protoParams := complianceproto.ComplianceParams{
		KycRequired:             params.KycRequired,
		AmlEnabled:              params.AmlEnabled,
		SanctionsCheckRequired:  params.SanctionsCheckRequired,
		TaxReportingEnabled:     params.TaxReportingEnabled,
		GdprComplianceEnabled:   params.GdprComplianceEnabled,
		MaxKycValidityDays:      params.MaxKycValidityDays,
		MaxAmlValidityDays:      params.MaxAmlValidityDays,
		MinKycTier:              params.MinKycTier,
		RequiredKycFields:       params.RequiredKycFields,
		RestrictedJurisdictions: params.RestrictedJurisdictions,
		AllowedJurisdictions:    params.AllowedJurisdictions,
	}

	return &complianceproto.GenesisState{
		Params:               protoParams,
		KycRecords:           []*complianceproto.KYCRecord{},
		AmlProfiles:          []*complianceproto.AMLProfile{},
		SuspiciousActivities: []*complianceproto.SuspiciousActivity{},
		MonitoringRules:      []*complianceproto.TransactionMonitoringRule{},
		TransactionAlerts:    []*complianceproto.TransactionAlert{},
		SanctionsResults:     []*complianceproto.SanctionsScreeningResult{},
		GdprConsents:         []*complianceproto.GDPRConsent{},
		GdprRequests:         []*complianceproto.GDPRDataRequest{},
		TaxReports:           []*complianceproto.TaxReport{},
	}
}
