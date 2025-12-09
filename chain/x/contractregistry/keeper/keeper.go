package keeper

import (
	"encoding/binary"

	storetypes "cosmossdk.io/store/types"
	"github.com/aequitas/aura/chain/x/contractregistry/types"
	pb "github.com/aequitas/aura/proto/aura/contractregistry/v1beta1"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper manages the state of the contractregistry module
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	authority  string
	compliance types.ComplianceKeeper
	vcKeeper   types.VCKeeper
	csKeeper   types.ConfidenceScoreKeeper
}

// NewKeeper creates a new Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority string,
) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		authority: authority,
	}
}

// SetComplianceKeeper sets the compliance keeper
func (k *Keeper) SetComplianceKeeper(keeper types.ComplianceKeeper) {
	k.compliance = keeper
}

// SetVCKeeper sets the VC keeper
func (k *Keeper) SetVCKeeper(keeper types.VCKeeper) {
	k.vcKeeper = keeper
}

// SetConfidenceScoreKeeper sets the confidence score keeper
func (k *Keeper) SetConfidenceScoreKeeper(keeper types.ConfidenceScoreKeeper) {
	k.csKeeper = keeper
}

// GetAuthority returns the governance module account address
func (k *Keeper) GetAuthority() string {
	return k.authority
}

// GetParams returns the current module parameters
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		// Return default params if not set
		return types.Params{
			AuditWarningDays:       90,
			MaxContractsPerCreator: 100,
		}
	}

	var protoParams pb.ContractRegistryParams
	k.cdc.MustUnmarshal(bz, &protoParams)

	// Convert proto params to types.Params
	return types.Params{
		AuditWarningDays:       protoParams.AuditWarningDays,
		MaxContractsPerCreator: protoParams.MaxContractsPerCreator,
	}
}

// SetParams sets new module parameters
func (k Keeper) SetParams(ctx sdk.Context, params interface{}) error {
	store := ctx.KVStore(k.storeKey)

	var protoParams *pb.ContractRegistryParams
	switch v := params.(type) {
	case types.Params:
		protoParams = &pb.ContractRegistryParams{
			AuditWarningDays:       v.AuditWarningDays,
			MaxContractsPerCreator: v.MaxContractsPerCreator,
		}
	case *types.Params:
		protoParams = &pb.ContractRegistryParams{
			AuditWarningDays:       v.AuditWarningDays,
			MaxContractsPerCreator: v.MaxContractsPerCreator,
		}
	case *pb.ContractRegistryParams:
		protoParams = v
	case pb.ContractRegistryParams:
		protoParams = &v
	default:
		// Fallback to default params
		protoParams = &pb.ContractRegistryParams{
			AuditWarningDays:       90,
			MaxContractsPerCreator: 100,
		}
	}

	bz := k.cdc.MustMarshal(protoParams)
	store.Set(types.ParamsKey, bz)
	return nil
}


// GetContractInfo retrieves contract info
func (k Keeper) GetContractInfo(ctx sdk.Context, contractAddr string) (pb.ContractInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.ContractInfoKey(contractAddr)
	bz := store.Get(key)
	if bz == nil {
		return pb.ContractInfo{}, false
	}

	var info pb.ContractInfo
	k.cdc.MustUnmarshal(bz, &info)
	return info, true
}

// SetContractInfo stores contract info
func (k Keeper) SetContractInfo(ctx sdk.Context, info *pb.ContractInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.ContractInfoKey(info.Address)
	bz := k.cdc.MustMarshal(info)
	store.Set(key, bz)
}

// GetContractMetrics retrieves contract metrics
func (k Keeper) GetContractMetrics(ctx sdk.Context, contractAddr string) (*pb.ContractMetrics, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.ContractMetricsKey(contractAddr)
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var metrics pb.ContractMetrics
	k.cdc.MustUnmarshal(bz, &metrics)
	return &metrics, true
}

// UpdateMetricsOnExecution updates metrics after contract execution
func (k Keeper) UpdateMetricsOnExecution(ctx sdk.Context, contractAddr string, gasUsed uint64, success bool) {
	metrics, found := k.GetContractMetrics(ctx, contractAddr)
	if !found {
		metrics = &pb.ContractMetrics{
			ContractAddress: contractAddr,
		}
	}
	metrics.TotalExecutions++
	if success {
		metrics.SuccessfulExecutions++
	} else {
		metrics.FailedExecutions++
	}
	metrics.TotalGasUsed += gasUsed

	// Calculate average gas per execution
	if metrics.TotalExecutions > 0 {
		metrics.AvgGasPerExecution = metrics.TotalGasUsed / metrics.TotalExecutions
	}

	store := ctx.KVStore(k.storeKey)
	key := types.ContractMetricsKey(contractAddr)
	bz := k.cdc.MustMarshal(metrics)
	store.Set(key, bz)
}

// IncrementMetricsCounter increments a specific metric counter
func (k Keeper) IncrementMetricsCounter(ctx sdk.Context, contractAddr string, counterType string) {
	metrics, found := k.GetContractMetrics(ctx, contractAddr)
	if !found {
		metrics = &pb.ContractMetrics{
			ContractAddress: contractAddr,
		}
	}
	switch counterType {
	case "rate_limit_violation":
		metrics.RateLimitViolations++
	case "compliance_failure":
		metrics.ComplianceFailures++
	}

	store := ctx.KVStore(k.storeKey)
	key := types.ContractMetricsKey(contractAddr)
	bz := k.cdc.MustMarshal(metrics)
	store.Set(key, bz)
}

// CheckRateLimit checks rate limiting
func (k Keeper) CheckRateLimit(ctx sdk.Context, contractAddr string, executor string, limit uint64) error {
	if limit == 0 {
		return nil
	}

	// Calculate the current time window
	blockTime := ctx.BlockTime().Unix()
	windowSize := int64(3600) // 1 hour windows
	windowStart := (blockTime / windowSize) * windowSize

	// Get current count for this window
	currentCount := k.GetRateLimitCount(ctx, contractAddr, executor, windowStart)

	// Check if limit exceeded
	if currentCount >= limit {
		return types.ErrRateLimitExceeded
	}

	return nil
}

// IncrementRateLimit increments the rate limit counter for a contract and user
// This should be called after successful execution to track usage
func (k Keeper) IncrementRateLimit(ctx sdk.Context, contractAddr, userAddr string) {
	// Calculate the current time window (e.g., hourly windows)
	blockTime := ctx.BlockTime().Unix()
	windowSize := int64(3600) // 1 hour windows
	windowStart := (blockTime / windowSize) * windowSize

	// Get current count for this window
	currentCount := k.GetRateLimitCount(ctx, contractAddr, userAddr, windowStart)

	// Increment count
	k.SetRateLimitCount(ctx, contractAddr, userAddr, windowStart, currentCount+1)
}

// GetRateLimitStatus returns the current rate limit status for a contract and user
// Returns: current count, remaining executions, and window reset timestamp
func (k Keeper) GetRateLimitStatus(ctx sdk.Context, contractAddr, userAddr string, limit uint64) (current, remaining uint64, reset int64) {
	// Calculate the current time window
	blockTime := ctx.BlockTime().Unix()
	windowSize := int64(3600) // 1 hour windows
	windowStart := (blockTime / windowSize) * windowSize
	resetTime := windowStart + windowSize

	// Get current count for this window
	current = k.GetRateLimitCount(ctx, contractAddr, userAddr, windowStart)

	// Calculate remaining
	if limit == 0 {
		// No limit set
		remaining = ^uint64(0) // Max uint64
	} else if current >= limit {
		remaining = 0
	} else {
		remaining = limit - current
	}

	return current, remaining, resetTime
}

// SetRateLimitCount sets the rate limit count for a specific window
func (k Keeper) SetRateLimitCount(ctx sdk.Context, contractAddr, userAddr string, window int64, count uint64) {
	store := ctx.KVStore(k.storeKey)
	key := types.RateLimitKey(contractAddr, userAddr, window)

	// Encode count as binary (8 bytes for uint64)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)

	store.Set(key, bz)
}

// GetRateLimitCount gets the current count for a specific window
// Returns 0 if no count exists for this window
func (k Keeper) GetRateLimitCount(ctx sdk.Context, contractAddr, userAddr string, window int64) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.RateLimitKey(contractAddr, userAddr, window)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}

	// Decode binary uint64
	return binary.BigEndian.Uint64(bz)
}

// CleanupOldRateLimits removes rate limit entries older than the retention period
// This should be called periodically (e.g., in EndBlocker) to prevent unbounded storage growth
func (k Keeper) CleanupOldRateLimits(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// Define retention period (e.g., 24 hours of windows)
	retentionPeriod := int64(24 * 3600) // 24 hours in seconds
	windowSize := int64(3600) // 1 hour windows

	// Calculate cutoff time
	blockTime := ctx.BlockTime().Unix()
	cutoffWindow := ((blockTime - retentionPeriod) / windowSize) * windowSize

	// Iterate over all rate limit entries
	iterator := storetypes.KVStorePrefixIterator(store, types.RateLimitPrefix)
	defer iterator.Close()

	keysToDelete := [][]byte{}
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		// Extract window timestamp from key
		// Key format: RateLimitKeyPrefix | contractAddr | "/" | userAddr | "/" | windowTimestamp (8 bytes)
		// Find the last 8 bytes which contain the timestamp
		if len(key) < len(types.RateLimitPrefix)+8 {
			continue
		}

		windowBytes := key[len(key)-8:]
		windowTimestamp := int64(binary.BigEndian.Uint64(windowBytes))

		// If this window is older than cutoff, mark for deletion
		if windowTimestamp < cutoffWindow {
			keysToDelete = append(keysToDelete, key)
		}
	}

	// Delete old entries
	for _, key := range keysToDelete {
		store.Delete(key)
	}
}

// RegisterContract registers a new contract
func (k Keeper) RegisterContract(ctx sdk.Context, info *pb.ContractInfo) error {
	// Check if contract already exists
	if _, found := k.GetContractInfo(ctx, info.Address); found {
		return types.ErrContractAlreadyExists
	}

	// Check max contracts per creator
	params := k.GetParams(ctx)
	creatorContracts := k.GetCreatorContracts(ctx, info.Creator)
	if params.MaxContractsPerCreator > 0 && uint64(len(creatorContracts)) >= params.MaxContractsPerCreator {
		return types.ErrTooManyContracts
	}

	// Set contract info
	k.SetContractInfo(ctx, info)

	// Add to creator index
	k.AddCreatorContract(ctx, info.Creator, info.Address)

	// Add to tag indexes
	if info.Metadata != nil {
		for _, tag := range info.Metadata.Tags {
			k.AddTagContract(ctx, tag, info.Address)
		}
	}

	// Initialize metrics
	metrics := &pb.ContractMetrics{
		ContractAddress: info.Address,
	}
	store := ctx.KVStore(k.storeKey)
	key := types.ContractMetricsKey(info.Address)
	bz := k.cdc.MustMarshal(metrics)
	store.Set(key, bz)

	return nil
}

// UpdateContractMetadata updates contract metadata
func (k Keeper) UpdateContractMetadata(ctx sdk.Context, contractAddr, signer string, metadata *pb.ContractMetadata) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check authorization (must be admin)
	if info.Admin != signer {
		return types.ErrNotContractAdmin
	}

	// Remove old tag indexes
	if info.Metadata != nil {
		for _, tag := range info.Metadata.Tags {
			k.RemoveTagContract(ctx, tag, contractAddr)
		}
	}

	// Update metadata
	info.Metadata = metadata
	k.SetContractInfo(ctx, &info)

	// Add new tag indexes
	if metadata != nil {
		for _, tag := range metadata.Tags {
			k.AddTagContract(ctx, tag, contractAddr)
		}
	}

	return nil
}

// UpdateSecurityPolicy updates contract security policy
func (k Keeper) UpdateSecurityPolicy(ctx sdk.Context, contractAddr, signer string, policy *pb.SecurityPolicy) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check authorization (must be admin)
	if info.Admin != signer {
		return types.ErrNotContractAdmin
	}

	// Update security policy
	info.SecurityPolicy = policy
	k.SetContractInfo(ctx, &info)

	return nil
}

// PauseContract pauses a contract
func (k Keeper) PauseContract(ctx sdk.Context, contractAddr, signer, reason string) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check if pause is allowed by security policy
	if info.SecurityPolicy != nil && !info.SecurityPolicy.AllowPause {
		return types.ErrInvalidRequest
	}

	// Check authorization (must be admin or governance)
	if info.Admin != signer && signer != k.authority {
		return types.ErrUnauthorized
	}

	// Update status
	info.Status = pb.ContractStatus_CONTRACT_STATUS_PAUSED
	k.SetContractInfo(ctx, &info)

	return nil
}

// UnpauseContract unpauses a contract
func (k Keeper) UnpauseContract(ctx sdk.Context, contractAddr, signer string) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check current status - can only unpause if paused
	if info.Status != pb.ContractStatus_CONTRACT_STATUS_PAUSED {
		return types.ErrInvalidRequest
	}

	// Check authorization (must be admin or governance)
	if info.Admin != signer && signer != k.authority {
		return types.ErrUnauthorized
	}

	// Update status
	info.Status = pb.ContractStatus_CONTRACT_STATUS_ACTIVE
	k.SetContractInfo(ctx, &info)

	return nil
}

// DeprecateContract deprecates a contract
func (k Keeper) DeprecateContract(ctx sdk.Context, contractAddr, signer, reason, migrationTarget string) error {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return types.ErrContractNotFound
	}

	// Check authorization (must be admin or governance)
	if info.Admin != signer && signer != k.authority {
		return types.ErrUnauthorized
	}

	// Update status
	info.Status = pb.ContractStatus_CONTRACT_STATUS_DEPRECATED
	// Note: MigrationTarget field not available in current proto definition
	// NOTE: Future enhancement - Regenerate proto to include migration_target field
	k.SetContractInfo(ctx, &info)

	return nil
}

// IsBlacklisted checks if a contract is blacklisted
func (k Keeper) IsBlacklisted(ctx sdk.Context, contractAddr string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.BlacklistKey(contractAddr)
	return store.Has(key)
}

// IsContractApproved checks if a contract is approved for execution
// Returns false if contract is paused or frozen, true otherwise (including deprecated contracts)
func (k Keeper) IsContractApproved(ctx sdk.Context, contractAddr string) bool {
	info, found := k.GetContractInfo(ctx, contractAddr)
	if !found {
		return false
	}

	// Contract is not approved if paused or frozen
	if info.Status == pb.ContractStatus_CONTRACT_STATUS_PAUSED ||
		info.Status == pb.ContractStatus_CONTRACT_STATUS_FROZEN {
		return false
	}

	return true
}

// IsContractRegistered returns true if contract exists in the registry
func (k Keeper) IsContractRegistered(ctx sdk.Context, contractAddr string) bool {
	_, found := k.GetContractInfo(ctx, contractAddr)
	return found
}

// GetCreatorContracts returns all contracts created by a specific creator
func (k Keeper) GetCreatorContracts(ctx sdk.Context, creator string) []*pb.ContractInfo {
	var contracts []*pb.ContractInfo
	store := ctx.KVStore(k.storeKey)

	// Iterate over creator index
	iterator := storetypes.KVStorePrefixIterator(store, types.CreatorContractsPrefix(creator))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract contract address from the index key
		// The key format is: CreatorContractsKeyPrefix + creator + '/' + contractAddr
		key := iterator.Key()
		prefixLen := len(types.CreatorContractsPrefix(creator))
		if len(key) <= prefixLen {
			continue
		}
		contractAddr := string(key[prefixLen:])

		// Retrieve the full contract info
		if info, found := k.GetContractInfo(ctx, contractAddr); found {
			contracts = append(contracts, &info)
		}
	}

	return contracts
}

// GetTagContracts returns all contracts with a specific tag
func (k Keeper) GetTagContracts(ctx sdk.Context, tag string) []*pb.ContractInfo {
	var contracts []*pb.ContractInfo
	store := ctx.KVStore(k.storeKey)

	// Iterate over tag index
	iterator := storetypes.KVStorePrefixIterator(store, types.TagContractsPrefix(tag))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Extract contract address from the index key
		// The key format is: TagContractsKeyPrefix + tag + '/' + contractAddr
		key := iterator.Key()
		prefixLen := len(types.TagContractsPrefix(tag))
		if len(key) <= prefixLen {
			continue
		}
		contractAddr := string(key[prefixLen:])

		// Retrieve the full contract info
		if info, found := k.GetContractInfo(ctx, contractAddr); found {
			contracts = append(contracts, &info)
		}
	}

	return contracts
}

// AddCreatorContract adds a contract to the creator index
func (k Keeper) AddCreatorContract(ctx sdk.Context, creator, contractAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreatorContractsKey(creator, contractAddr)
	// Store a marker value (empty byte to indicate presence)
	store.Set(key, []byte{0x01})
}

// AddTagContract adds a contract to a tag index
func (k Keeper) AddTagContract(ctx sdk.Context, tag, contractAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.TagContractsKey(tag, contractAddr)
	// Store a marker value (empty byte to indicate presence)
	store.Set(key, []byte{0x01})
}

// RemoveTagContract removes a contract from a tag index
func (k Keeper) RemoveTagContract(ctx sdk.Context, tag, contractAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.TagContractsKey(tag, contractAddr)
	store.Delete(key)
}

// DeleteContractInfo removes a contract from storage
func (k Keeper) DeleteContractInfo(ctx sdk.Context, contractAddr string) {
	store := ctx.KVStore(k.storeKey)
	key := types.ContractInfoKey(contractAddr)
	store.Delete(key)
}

// GetAllContracts returns all registered contracts
func (k Keeper) GetAllContracts(ctx sdk.Context) []*pb.ContractInfo {
	var contracts []*pb.ContractInfo
	store := ctx.KVStore(k.storeKey)

	// Iterate over all contract info entries
	iterator := storetypes.KVStorePrefixIterator(store, types.ContractInfoPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info pb.ContractInfo
		k.cdc.MustUnmarshal(iterator.Value(), &info)
		contracts = append(contracts, &info)
	}

	return contracts
}
