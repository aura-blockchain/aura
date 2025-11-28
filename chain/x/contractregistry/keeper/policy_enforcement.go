package keeper

import (
	"context"
	"fmt"

	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EnforceSecurityPolicy enforces security policy before contract execution
func (k Keeper) EnforceSecurityPolicy(ctx sdk.Context, contractAddr, executor string, gasLimit uint64) error {
	// Get contract info
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check if contract is active
	if info.Status != pb.ContractStatus_CONTRACT_STATUS_ACTIVE &&
		info.Status != pb.ContractStatus_CONTRACT_STATUS_DEPRECATED {
		return types.ErrContractNotActive
	}

	// Check if contract is blacklisted
	if k.IsBlacklisted(ctx, contractAddr) {
		return types.ErrContractBlacklisted
	}

	// Check security policy
	policy := info.SecurityPolicy
	if policy != nil {
		// Enforce rate limiting
		if policy.RateLimitPerUser > 0 {
			if err := k.CheckRateLimit(ctx, contractAddr, executor, policy.RateLimitPerUser); err != nil {
				k.IncrementMetricsCounter(ctx, contractAddr, "rate_limit_violation")
				return err
			}
		}

		// Enforce gas limits
		if policy.MaxGasPerTx > 0 && gasLimit > policy.MaxGasPerTx {
			return types.ErrGasLimitExceeded
		}

		// Check blacklist
		for _, addr := range policy.BlacklistedAddresses {
			if addr == executor {
				return types.ErrBlacklisted
			}
		}

		// Check whitelist (if non-empty, executor must be on it)
		if len(policy.WhitelistedAddresses) > 0 {
			found := false
			for _, addr := range policy.WhitelistedAddresses {
				if addr == executor {
					found = true
					break
				}
			}
			if !found {
				return types.ErrNotWhitelisted
			}
		}
	}

	// Check metadata-based requirements
	metadata := info.Metadata
	if metadata != nil {
		// Enforce KYC requirements
		if metadata.RequiredKycLevel > 0 && k.compliance != nil {
			level, err := k.compliance.GetKYCLevel(ctx, executor)
			if err != nil || level < metadata.RequiredKycLevel {
				return types.ErrKYCRequired
			}
		}

		// Check sanctions screening
		if metadata.CheckSanctions && k.compliance != nil {
			isSanctioned, err := k.compliance.ScreenForSanctions(ctx, executor)
			if err != nil {
				return err
			}
			if isSanctioned {
				return types.ErrSanctioned
			}
		}

		// Enforce VC requirements
		if metadata.RequiresVc && k.vcKeeper != nil {
			if len(metadata.RequiredVcTypes) > 0 {
				hasRequiredVC := false
				for _, vcType := range metadata.RequiredVcTypes {
					if k.vcKeeper.HasVC(context.Background(), executor, vcType) {
						hasRequiredVC = true
						break
					}
				}
				if !hasRequiredVC {
					return types.ErrVCRequired
				}
			}
		}

		// Enforce confidence score requirements
		if metadata.MinConfidenceScore > 0 && k.csKeeper != nil {
			score, found := k.csKeeper.GetUserScore(executor)
			if !found || score < metadata.MinConfidenceScore {
				return types.ErrLowConfidenceScore
			}
		}
	}

	// Check compliance requirements
	compliance := info.Compliance
	if compliance != nil && k.compliance != nil {
		// Enforce KYC
		if compliance.EnforceKyc {
			level, err := k.compliance.GetKYCLevel(ctx, executor)
			if err != nil || level < compliance.MinKycLevel {
				return types.ErrKYCRequired
			}
		}

		// Enforce sanctions screening
		if compliance.EnforceSanctionsCheck {
			isSanctioned, err := k.compliance.ScreenForSanctions(ctx, executor)
			if err != nil {
				return err
			}
			if isSanctioned {
				return types.ErrSanctioned
			}
		}
	}

	return nil
}

// RecordPolicyViolation records a policy violation in the audit trail
func (k Keeper) RecordPolicyViolation(ctx sdk.Context, contractAddr, executor, violationType, details string) {
	// Create audit entry
	entry := &types.AuditEntry{
		ContractAddress: contractAddr,
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		Action:          "POLICY_VIOLATION",
		Actor:           executor,
		Details:         fmt.Sprintf("Type: %s, Details: %s", violationType, details),
		Success:         false,
	}

	k.AddAuditEntry(ctx, entry)

	// Increment metrics
	k.IncrementMetricsCounter(ctx, contractAddr, "compliance_failure")

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"contract_policy_violation",
			sdk.NewAttribute("contract_address", contractAddr),
			sdk.NewAttribute("executor", executor),
			sdk.NewAttribute("violation_type", violationType),
			sdk.NewAttribute("details", details),
		),
	)
}

// ValidateComplianceRequirements validates compliance configuration
func (k Keeper) ValidateComplianceRequirements(ctx sdk.Context, reqs *pb.ComplianceRequirements) error {
	if reqs == nil {
		return nil
	}

	// Validate KYC levels
	if reqs.MinKycLevel > 5 {
		return types.ErrInvalidParams
	}

	// Note: RestrictedJurisdictions field not in current proto
	// Skipping this validation for now

	return nil
}

// CheckJurisdictionRestrictions checks if user's jurisdiction is restricted
func (k Keeper) CheckJurisdictionRestrictions(ctx sdk.Context, contractAddr, userAddr string) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Note: RestrictedJurisdictions field not in current proto definition
	// For now, just do basic compliance check
	if info.Compliance == nil {
		return nil
	}

	// Get user's jurisdiction from compliance module
	if k.compliance != nil {
		// Note: This would require extending the compliance keeper interface
		// For now, we just check if the user passes basic compliance
		_, err := k.compliance.GetKYCLevel(ctx, userAddr)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateComplianceRequirements updates contract compliance requirements
func (k Keeper) UpdateComplianceRequirements(ctx sdk.Context, contractAddr, admin string, reqs *pb.ComplianceRequirements) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	if info.Admin != admin {
		return types.ErrNotContractAdmin
	}

	if err := k.ValidateComplianceRequirements(ctx, reqs); err != nil {
		return err
	}

	info.Compliance = reqs
	k.SetContractInfo(ctx, &info)

	// Record in audit trail
	k.AddAuditEntry(ctx, &types.AuditEntry{
		ContractAddress: contractAddr,
		Timestamp:       timestamppb.New(ctx.BlockTime()),
		Action:          "UPDATE_COMPLIANCE_REQUIREMENTS",
		Actor:           admin,
		Success:         true,
	})

	return nil
}
