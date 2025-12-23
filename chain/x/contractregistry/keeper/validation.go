package keeper

import (
	"fmt"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateContractExecution validates that a contract execution meets all requirements
func (k Keeper) ValidateContractExecution(ctx sdk.Context, contractAddr, sender string, gasLimit uint64) error {
	// 1. Check if contract is registered
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// 2. Check contract status
	switch info.Status {
	case pb.ContractStatus_CONTRACT_STATUS_PAUSED:
		return types.ErrContractPaused
	case pb.ContractStatus_CONTRACT_STATUS_FROZEN:
		return types.ErrContractFrozen
	case pb.ContractStatus_CONTRACT_STATUS_UNSPECIFIED:
		return types.ErrInvalidRequest
	}

	// 3. Check blacklist/whitelist
	if err := k.checkAddressLists(sender, &info.SecurityPolicy); err != nil {
		return fmt.Errorf("error in ValidateContractExecution for ErrInvalidRequest: %w", err)
	}

	// 4. Check gas limit
	if info.SecurityPolicy.MaxGasPerTx > 0 && gasLimit > info.SecurityPolicy.MaxGasPerTx {
		return types.ErrGasLimitExceeded
	}

	// 5. Check rate limits
	if info.SecurityPolicy.RateLimitPerUser > 0 {
		if err := k.CheckRateLimit(ctx, contractAddr, sender, info.SecurityPolicy.RateLimitPerUser); err != nil {
			k.IncrementMetricsCounter(ctx, contractAddr, "rate_limit_violation")
			return fmt.Errorf("error in ValidateContractExecution: %w", err)
		}
	}

	// 6. Check compliance requirements
	if err := k.checkCompliance(ctx, sender, &info.Compliance, &info.Metadata); err != nil {
		k.IncrementMetricsCounter(ctx, contractAddr, "compliance_failure")
		return fmt.Errorf("error in ValidateContractExecution: %w", err)
	}

	return nil
}

// checkAddressLists checks if sender is blacklisted or not whitelisted
func (k Keeper) checkAddressLists(sender string, policy *pb.SecurityPolicy) error {
	if policy == nil {
		return nil
	}

	// Check blacklist
	for _, addr := range policy.BlacklistedAddresses {
		if addr == sender {
			return types.ErrBlacklisted
		}
	}

	// Check whitelist (if non-empty, sender must be on it)
	if len(policy.WhitelistedAddresses) > 0 {
		found := false
		for _, addr := range policy.WhitelistedAddresses {
			if addr == sender {
				found = true
				break
			}
		}
		if !found {
			return types.ErrNotWhitelisted
		}
	}

	return nil
}

// checkCompliance validates compliance requirements
func (k Keeper) checkCompliance(ctx sdk.Context, sender string, compliance *pb.ComplianceRequirements, metadata *pb.ContractMetadata) error {
	if compliance == nil || metadata == nil {
		return nil
	}

	// Check KYC requirements
	if compliance.EnforceKyc || metadata.RequiredKycLevel > 0 {
		requiredLevel := compliance.MinKycLevel
		if metadata.RequiredKycLevel > requiredLevel {
			requiredLevel = metadata.RequiredKycLevel
		}

		if k.compliance != nil {
			level, err := k.compliance.GetKYCLevel(ctx, sender)
			if err != nil || level < requiredLevel {
				return types.ErrKYCRequired
			}
		}
	}

	// Check sanctions screening
	if compliance.EnforceSanctionsCheck || metadata.CheckSanctions {
		if k.compliance != nil {
			sanctioned, err := k.compliance.ScreenForSanctions(ctx, sender)
			if err != nil || sanctioned {
				return types.ErrSanctionsCheckFailed
			}
		}
	}

	// Check VC requirements
	if metadata.RequiresVc && len(metadata.RequiredVcTypes) > 0 {
		if k.vcKeeper != nil {
			for _, vcType := range metadata.RequiredVcTypes {
				if !k.vcKeeper.HasVC(ctx, sender, vcType) {
					return fmt.Errorf("%w: %s", types.ErrMissingVC, vcType)
				}
			}
		}
	}

	// Check confidence score
	if metadata.MinConfidenceScore > 0 {
		if k.csKeeper != nil {
			score, exists := k.csKeeper.GetUserScore(ctx, sender)
			if !exists || score < metadata.MinConfidenceScore {
				return types.ErrInsufficientCS
			}
		}
	}

	return nil
}

// ValidateContractRegistration validates contract registration request
func (k Keeper) ValidateContractRegistration(ctx sdk.Context, msg *types.MsgRegisterContract) error {
	// Validate contract address
	if msg.ContractAddress == "" {
		return types.ErrInvalidContractAddress
	}

	// Validate code ID
	if msg.CodeId == 0 {
		return types.ErrInvalidCodeID
	}

	// Validate creator and admin
	if msg.Creator == "" || msg.Admin == "" {
		return types.ErrInvalidRequest
	}

	// Validate metadata - must be provided and valid
	if msg.Metadata == nil {
		return types.ErrInvalidMetadata
	}
	if err := k.ValidateMetadataUpdate(convertToProtoMetadata(msg.Metadata)); err != nil {
		return fmt.Errorf("error in ValidateContractRegistration for Validate: %w", err)
	}

	// Validate security policy if provided
	if msg.SecurityPolicy != nil {
		if err := k.ValidateSecurityPolicyUpdate(convertToProtoSecurityPolicy(msg.SecurityPolicy)); err != nil {
			return fmt.Errorf("error in ValidateContractRegistration for ErrInvalidMetadata: %w", err)
		}
	}

	return nil
}

// convertToProtoMetadata converts types.ContractMetadata to pb.ContractMetadata
func convertToProtoMetadata(meta *types.ContractMetadata) *pb.ContractMetadata {
	if meta == nil {
		return nil
	}
	return &pb.ContractMetadata{
		Name:        meta.Name,
		Description: meta.Description,
		Version:     meta.Version,
		Tags:        meta.Tags,
	}
}

// convertToProtoSecurityPolicy converts types.SecurityPolicy to pb.SecurityPolicy
func convertToProtoSecurityPolicy(policy *types.SecurityPolicy) *pb.SecurityPolicy {
	if policy == nil {
		return nil
	}
	return &pb.SecurityPolicy{
		AllowPause:           policy.AllowPause,
		MaxGasPerTx:          policy.MaxGasPerExecution,
		RateLimitPerUser:     policy.RateLimitPerUser,
		BlacklistedAddresses: policy.AllowedExecutors, // Use AllowedExecutors as a substitute
		WhitelistedAddresses: policy.AllowedExecutors,
	}
}

// ValidateMetadataUpdate validates metadata update request
func (k Keeper) ValidateMetadataUpdate(metadata *pb.ContractMetadata) error {
	if metadata == nil {
		return types.ErrInvalidMetadata
	}

	if metadata.Name == "" {
		return types.ErrInvalidMetadata
	}

	// Validate required VC types format
	for _, vcType := range metadata.RequiredVcTypes {
		if vcType == "" {
			return types.ErrInvalidMetadata
		}
	}

	// Validate confidence score range
	if metadata.MinConfidenceScore > 100 {
		return types.ErrInvalidMetadata
	}

	// Validate KYC level range (0-3)
	if metadata.RequiredKycLevel > 3 {
		return types.ErrInvalidMetadata
	}

	return nil
}

// ValidateSecurityPolicyUpdate validates security policy update
func (k Keeper) ValidateSecurityPolicyUpdate(policy *pb.SecurityPolicy) error {
	if policy == nil {
		return types.ErrInvalidSecurityPolicy
	}

	// Validate gas limits
	if policy.MaxGasPerTx > 50000000 {
		return types.ErrInvalidSecurityPolicy
	}

	// Validate rate limits
	if policy.RateLimitPerUser > 10000 {
		return types.ErrInvalidSecurityPolicy
	}

	// Validate address lists
	for _, addr := range policy.BlacklistedAddresses {
		if addr == "" {
			return types.ErrInvalidSecurityPolicy
		}
	}

	for _, addr := range policy.WhitelistedAddresses {
		if addr == "" {
			return types.ErrInvalidSecurityPolicy
		}
	}

	return nil
}
